package utils

import (
	"strings"
	"time"

	// Embed IANA zone database so LoadLocation works without system zoneinfo
	// (Windows hosts and minimal containers).
	_ "time/tzdata"
)

// IsValidTimezone returns true if tz is a non-empty IANA timezone that Go can load.
func IsValidTimezone(tz string) bool {
	tz = strings.TrimSpace(tz)
	if tz == "" {
		return false
	}
	_, err := time.LoadLocation(tz)
	return err == nil
}

// ValidItemsPerPage returns true for supported pagination sizes.
func ValidItemsPerPage(n int) bool {
	switch n {
	case 10, 15, 25, 50:
		return true
	default:
		return false
	}
}
