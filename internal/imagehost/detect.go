package imagehost

import (
	"bytes"
	"fmt"
	"net/http"
	"strings"
)

const (
	TypeJPEG = "image/jpeg"
	TypePNG  = "image/png"
	TypeGIF  = "image/gif"
	TypeWebP = "image/webp"
)

var allowedTypes = map[string]string{
	TypeJPEG: ".jpg",
	TypePNG:  ".png",
	TypeGIF:  ".gif",
	TypeWebP: ".webp",
}

// ErrNotImage is returned when uploaded bytes are not a supported still image.
var ErrNotImage = fmt.Errorf("file is not a supported image (jpeg, png, gif, or webp)")

// DetectImage returns the canonical content type for a supported image.
// SVG and other types are rejected to avoid stored XSS.
func DetectImage(data []byte) (string, error) {
	if len(data) < 12 {
		return "", ErrNotImage
	}
	if t := sniffWebP(data); t != "" {
		return t, nil
	}
	t := http.DetectContentType(data)
	switch t {
	case TypeJPEG, TypePNG, TypeGIF:
		return t, nil
	default:
		return "", ErrNotImage
	}
}

func sniffWebP(data []byte) string {
	if len(data) < 12 {
		return ""
	}
	if bytes.Equal(data[:4], []byte("RIFF")) && bytes.Equal(data[8:12], []byte("WEBP")) {
		return TypeWebP
	}
	return ""
}

// ExtForContentType returns the file extension including the leading dot.
func ExtForContentType(contentType string) string {
	if ext, ok := allowedTypes[strings.ToLower(strings.TrimSpace(contentType))]; ok {
		return ext
	}
	return ""
}

// AllowedContentType reports whether contentType is a hosted image type.
func AllowedContentType(contentType string) bool {
	_, ok := allowedTypes[strings.ToLower(strings.TrimSpace(contentType))]
	return ok
}
