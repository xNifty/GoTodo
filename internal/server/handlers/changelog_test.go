package handlers

import (
	"strings"
	"testing"
)

func TestRenderMarkdownAutolinksBareURLs(t *testing.T) {
	md := `## What's Changed
* Feature by @user in https://github.com/owner/repo/pull/123
* Commit https://github.com/owner/repo/commit/abc123
**Full Changelog**: https://github.com/owner/repo/compare/v1.0.0...v1.1.0
`
	html := renderMarkdown(md)
	for _, url := range []string{
		"https://github.com/owner/repo/pull/123",
		"https://github.com/owner/repo/commit/abc123",
		"https://github.com/owner/repo/compare/v1.0.0...v1.1.0",
	} {
		want := `<a target="_blank" rel="noopener noreferrer" href="` + url + `"`
		if !strings.Contains(html, want) {
			t.Fatalf("expected clickable link for %s, got:\n%s", url, html)
		}
	}
}

func TestRenderMarkdownKeepsExplicitMarkdownLinks(t *testing.T) {
	html := renderMarkdown(`See [#456](https://github.com/owner/repo/pull/456)`)
	want := `<a target="_blank" rel="noopener noreferrer" href="https://github.com/owner/repo/pull/456">#456</a>`
	if !strings.Contains(html, want) {
		t.Fatalf("expected markdown link rendered, got:\n%s", html)
	}
}
