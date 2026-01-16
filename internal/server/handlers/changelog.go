package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"GoTodo/internal/config"
	srvutils "GoTodo/internal/server/utils"

	"github.com/yuin/goldmark"
)

// ChangelogEntry is the public structure returned to the client
type ChangelogEntry struct {
	Version string   `json:"version"`
	Date    string   `json:"date"`
	Title   string   `json:"title"`
	Notes   []string `json:"notes"`
	Html    string   `json:"html,omitempty"`
}

// In-memory fallback cache
type memItem struct {
	data   string
	etag   string
	expiry time.Time
}

var memCache = struct {
	m  map[string]memItem
	mu sync.RWMutex
}{m: make(map[string]memItem)}

// ChangelogHandler serves the changelog JSON; it will attempt to pull from
// GitHub releases when GITHUB_REPO is set (owner/repo). If that fails, it
// falls back to config/changelog.json.
func ChangelogHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Respect config toggle
	if !config.Cfg.ShowChangelog {
		http.NotFound(w, r)
		return
	}

	// If GITHUB_REPO is configured, try fetching releases
	repo := strings.TrimSpace(os.Getenv("GITHUB_REPO"))
	if repo != "" {
		if entries, err := fetchFromGitHub(repo); err == nil {
			respondJSON(w, entries)
			return
		}
		// else fall through to local file
	}

	// Local fallback
	cfgPath := filepath.Join("config", "changelog.json")
	data, err := os.ReadFile(cfgPath)
	if err != nil {
		respondJSON(w, []ChangelogEntry{})
		return
	}

	// Validate and return
	var v []ChangelogEntry
	if err := json.Unmarshal(data, &v); err != nil {
		respondJSON(w, []ChangelogEntry{})
		return
	}
	// Render HTML for any local entries (join notes into markdown list and convert)
	for i := range v {
		if v[i].Html == "" {
			if len(v[i].Notes) > 0 {
				md := ""
				for _, n := range v[i].Notes {
					md += "- " + n + "\n"
				}
				v[i].Html = renderMarkdown(md)
			}
		}
	}
	respondJSON(w, v)
}

func respondJSON(w http.ResponseWriter, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(v)
}

// fetchFromGitHub fetches releases from the GitHub API and maps them to ChangelogEntry
func fetchFromGitHub(repo string) ([]ChangelogEntry, error) {
	// Use Redis caching (with ETag) if available, otherwise in-memory TTL cache.
	ctx := context.Background()
	dataKey := fmt.Sprintf("changelog:data:%s", repo)
	etagKey := fmt.Sprintf("changelog:etag:%s", repo)

	var cachedJSON string
	var cachedETag string

	// Try Redis first
	if srvutils.RedisClient != nil {
		if v, err := srvutils.RedisClient.Get(ctx, dataKey).Result(); err == nil {
			cachedJSON = v
		}
		if e, err := srvutils.RedisClient.Get(ctx, etagKey).Result(); err == nil {
			cachedETag = e
		}
	} else {
		// In-memory fallback
		memCache.mu.RLock()
		if it, ok := memCache.m[repo]; ok && time.Now().Before(it.expiry) {
			cachedJSON = it.data
			cachedETag = it.etag
		}
		memCache.mu.RUnlock()
	}

	// repo expected as owner/repo
	url := fmt.Sprintf("https://api.github.com/repos/%s/releases", repo)
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	// Optionally use token for higher rate limits
	token := strings.TrimSpace(os.Getenv("GITHUB_TOKEN"))
	if token != "" {
		req.Header.Set("Authorization", "token "+token)
	}
	req.Header.Set("Accept", "application/vnd.github.v3+json")
	if cachedETag != "" {
		req.Header.Set("If-None-Match", cachedETag)
	}

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		// On network error, if cached JSON exists, return cached
		if cachedJSON != "" {
			var cached []ChangelogEntry
			_ = json.Unmarshal([]byte(cachedJSON), &cached)
			return cached, nil
		}
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotModified {
		// 304 — use cached data
		if cachedJSON != "" {
			var cached []ChangelogEntry
			if err := json.Unmarshal([]byte(cachedJSON), &cached); err == nil {
				return cached, nil
			}
		}
		return nil, fmt.Errorf("received 304 but no cached data")
	}

	if resp.StatusCode != http.StatusOK {
		io.Copy(io.Discard, resp.Body)
		// If we have cached payload, return it instead of failing
		if cachedJSON != "" {
			var cached []ChangelogEntry
			_ = json.Unmarshal([]byte(cachedJSON), &cached)
			return cached, nil
		}
		return nil, fmt.Errorf("github API returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		if cachedJSON != "" {
			var cached []ChangelogEntry
			_ = json.Unmarshal([]byte(cachedJSON), &cached)
			return cached, nil
		}
		return nil, err
	}

	// Minimal struct to decode releases
	var releases []struct {
		TagName     string `json:"tag_name"`
		Name        string `json:"name"`
		PublishedAt string `json:"published_at"`
		Body        string `json:"body"`
		Draft       bool   `json:"draft"`
		Prerelease  bool   `json:"prerelease"`
	}

	if err := json.Unmarshal(body, &releases); err != nil {
		if cachedJSON != "" {
			var cached []ChangelogEntry
			_ = json.Unmarshal([]byte(cachedJSON), &cached)
			return cached, nil
		}
		return nil, err
	}

	out := make([]ChangelogEntry, 0, len(releases))
	for _, r := range releases {
		if r.Draft {
			continue
		}
		date := r.PublishedAt
		// Trim time portion if present
		if strings.Contains(date, "T") {
			if t, err := time.Parse(time.RFC3339, date); err == nil {
				date = t.Format("2006-01-02")
			}
		}
		title := r.Name
		if title == "" {
			title = r.TagName
		}
		// Render the full markdown body from GitHub releases to HTML
		notes := parseNotesFromBody(r.Body)
		html := renderMarkdown(r.Body)
		out = append(out, ChangelogEntry{
			Version: r.TagName,
			Date:    date,
			Title:   title,
			Notes:   notes,
			Html:    html,
		})
	}

	// Marshal final payload and cache it with ETag
	finalB, _ := json.Marshal(out)
	newETag := resp.Header.Get("ETag")
	// Store in Redis if available
	if srvutils.RedisClient != nil {
		// cache for 10 minutes
		_ = srvutils.RedisClient.Set(ctx, dataKey, string(finalB), 10*time.Minute).Err()
		if newETag != "" {
			_ = srvutils.RedisClient.Set(ctx, etagKey, newETag, 10*time.Minute).Err()
		}
	} else {
		memCache.mu.Lock()
		memCache.m[repo] = memItem{data: string(finalB), etag: newETag, expiry: time.Now().Add(10 * time.Minute)}
		memCache.mu.Unlock()
	}

	return out, nil
}

func parseNotesFromBody(body string) []string {
	if strings.TrimSpace(body) == "" {
		return nil
	}
	lines := strings.Split(body, "\n")
	notes := make([]string, 0, len(lines))
	for _, l := range lines {
		s := strings.TrimSpace(l)
		if s == "" {
			continue
		}
		// strip common bullet markers
		if strings.HasPrefix(s, "- ") || strings.HasPrefix(s, "* ") || strings.HasPrefix(s, "• ") {
			if len(s) > 2 {
				s = strings.TrimSpace(s[2:])
			} else {
				s = ""
			}
		}
		if s != "" {
			notes = append(notes, s)
		}
	}
	return notes
}

// renderMarkdown converts markdown text to HTML using goldmark.
func renderMarkdown(md string) string {
	if strings.TrimSpace(md) == "" {
		return ""
	}
	var buf bytes.Buffer
	if err := goldmark.Convert([]byte(md), &buf); err != nil {
		return ""
	}
	return buf.String()
}

// PreloadChangelog attempts to fetch releases once (used at startup)
// It populates the Redis or in-memory cache via fetchFromGitHub.
func PreloadChangelog() error {
	if !config.Cfg.ShowChangelog {
		return nil
	}
	repo := strings.TrimSpace(os.Getenv("GITHUB_REPO"))
	if repo == "" {
		return nil
	}
	_, err := fetchFromGitHub(repo)
	return err
}

// ChangelogPageHandler renders a full HTML page showing all changelog entries
func ChangelogPageHandler(w http.ResponseWriter, r *http.Request) {
	if !config.Cfg.ShowChangelog {
		http.NotFound(w, r)
		return
	}

	var entries []ChangelogEntry
	repo := strings.TrimSpace(os.Getenv("GITHUB_REPO"))
	if repo != "" {
		if e, err := fetchFromGitHub(repo); err == nil {
			entries = e
		}
	}
	// If entries empty, try local file fallback
	if len(entries) == 0 {
		cfgPath := filepath.Join("config", "changelog.json")
		if data, err := os.ReadFile(cfgPath); err == nil {
			_ = json.Unmarshal(data, &entries)
			for i := range entries {
				if entries[i].Html == "" {
					if len(entries[i].Notes) > 0 {
						// Join notes into markdown list and render to HTML so local entries match GitHub rendering
						md := ""
						for _, n := range entries[i].Notes {
							md += "- " + n + "\n"
						}
						entries[i].Html = renderMarkdown(md)
					}
				}
			}
		}
	}

	ctx := map[string]interface{}{"Entries": entries}
	_ = srvutils.RenderTemplate(w, r, "changelog_page.html", ctx)
}
