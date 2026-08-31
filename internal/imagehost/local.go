package imagehost

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// LocalStore writes images to a directory on disk.
type LocalStore struct {
	Dir        string
	PublicBase string
}

// NewLocalStore creates a disk-backed store. dir is created if missing.
func NewLocalStore(dir, publicBase string) (*LocalStore, error) {
	dir = strings.TrimSpace(dir)
	if dir == "" {
		dir = DefaultLocalPath
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, fmt.Errorf("create upload directory: %w", err)
	}
	abs, err := filepath.Abs(dir)
	if err != nil {
		return nil, fmt.Errorf("resolve upload directory: %w", err)
	}
	return &LocalStore{Dir: abs, PublicBase: strings.TrimRight(strings.TrimSpace(publicBase), "/")}, nil
}

// Put writes obj to disk and returns the public URL.
func (s *LocalStore) Put(ctx context.Context, obj Object) (string, error) {
	if err := ctx.Err(); err != nil {
		return "", err
	}
	if !SafeObjectKey(obj.Key) {
		return "", fmt.Errorf("invalid object key")
	}
	dest := filepath.Join(s.Dir, obj.Key)
	if err := os.WriteFile(dest, obj.Data, 0o644); err != nil {
		return "", fmt.Errorf("write upload: %w", err)
	}
	return JoinPublicURL(s.PublicBase, obj.Key), nil
}

// Read returns the bytes for a previously stored local object.
func (s *LocalStore) Read(key string) ([]byte, string, error) {
	if !SafeObjectKey(key) {
		return nil, "", fmt.Errorf("invalid object key")
	}
	dest := filepath.Join(s.Dir, filepath.Base(key))
	data, err := os.ReadFile(dest)
	if err != nil {
		return nil, "", err
	}
	ct, err := DetectImage(data)
	if err != nil {
		// Serve by extension if magic-byte detection fails for a stored file.
		ct = contentTypeForExt(filepath.Ext(key))
		if ct == "" {
			return nil, "", err
		}
	}
	return data, ct, nil
}

func contentTypeForExt(ext string) string {
	switch strings.ToLower(ext) {
	case ".jpg", ".jpeg":
		return TypeJPEG
	case ".png":
		return TypePNG
	case ".gif":
		return TypeGIF
	case ".webp":
		return TypeWebP
	default:
		return ""
	}
}
