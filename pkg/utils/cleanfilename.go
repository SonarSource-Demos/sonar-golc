package utils

import (
	"strings"
)

func CleanFileName(originalName string) string {
	index := strings.Index(originalName, "gcloc-extract")
	if index != -1 {

		endIndex := strings.Index(originalName[index:], "/")
		if endIndex != -1 {
			endIndex += index
			return originalName[endIndex+1:]
		}
	}
	return originalName
}
