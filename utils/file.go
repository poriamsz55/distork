package utils

import (
	"fmt"
	"path/filepath"
	"strings"
)

func SanitizeFileName(fileName string, copyCount int) string {
	extension := filepath.Ext(fileName)
	name := strings.TrimSuffix(filepath.Base(fileName), extension)

	if copyCount == 0 {
		return fmt.Sprintf("%s%s", name, extension)
	}
	return fmt.Sprintf("%s (%d)%s", name, copyCount, extension)
}
