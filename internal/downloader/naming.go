package downloader

import (
	"fmt"
	"net/url"
	"path"
	"strings"
	"time"
)

// FileNameFromURL extracts a sensible filename from the URL.
func FileNameFromURL(raw string) string {
	u, err := url.Parse(raw)
	if err != nil {
		return fallbackName()
	}
	base := path.Base(u.Path)
	if base == "." || base == "/" || base == "" {
		return fallbackName()
	}
	base = strings.TrimSuffix(base, "/")
	if base == "" {
		return fallbackName()
	}
	return sanitize(base)
}

func fallbackName() string {
	return fmt.Sprintf("download_%d.bin", time.Now().UnixNano())
}

func sanitize(in string) string {
	if idx := strings.Index(in, "?"); idx >= 0 {
		in = in[:idx]
	}
	in = strings.TrimSpace(in)
	if in == "" {
		return fallbackName()
	}
	// Replace path separators just in case
	in = strings.ReplaceAll(in, "\\", "-")
	in = strings.ReplaceAll(in, "/", "-")
	return in
}
