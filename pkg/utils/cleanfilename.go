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
	count := 0
	thirdSlashIndex := -1

	for i := len(originalName) - 1; i >= 0; i-- {
		if originalName[i] == '/' {
			count++
			if count == 3 {
				thirdSlashIndex = i
				break
			}
		}
	}

	if thirdSlashIndex != -1 {
		return originalName[thirdSlashIndex+1:]
	}
	return originalName
}
