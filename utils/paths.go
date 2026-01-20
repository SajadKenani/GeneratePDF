package utils

import (
	"os"
	"path/filepath"
	"strings"
)

func UploadsDir() string {
	p := strings.TrimSpace(os.Getenv("UPLOADS_DIR"))
	if p == "" {
		p = "./uploads"
	}
	if abs, err := filepath.Abs(p); err == nil {
		return abs
	}
	return p
}

func CompressedDir() string {
	p := strings.TrimSpace(os.Getenv("COMPRESSED_DIR"))
	if p == "" {
		p = "./compressed"
	}
	if abs, err := filepath.Abs(p); err == nil {
		return abs
	}
	return p
}
