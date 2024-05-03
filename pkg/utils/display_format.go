package utils

import "fmt"

func FormatSize(size int64) string {
	const (
		byteSize = 1.0
		kiloSize = 1024.0
		megaSize = 1024.0 * kiloSize
		gigaSize = 1024.0 * megaSize
	)

	switch {
	case size < kiloSize:
		return fmt.Sprintf("%d B", size)
	case size < megaSize:
		return fmt.Sprintf("%.2f KB", float64(size)/kiloSize)
	case size < gigaSize:
		return fmt.Sprintf("%.2f MB", float64(size)/megaSize)
	default:
		return fmt.Sprintf("%.2f GB", float64(size)/gigaSize)
	}
}

func FormatCodeLines(numLines float64) string {
	if numLines >= 1000000 {
		return fmt.Sprintf("%.2fM", numLines/1000000)
	} else if numLines >= 1000 {
		return fmt.Sprintf("%.2fK", numLines/1000)
	} else {
		return fmt.Sprintf("%.0f", numLines)
	}
}
