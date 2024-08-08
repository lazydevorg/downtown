package main

import "fmt"

const (
	_  = iota
	KB = 1 << (10 * iota)
	MB
	GB
)

func HumanizeSize(bytes int64) string {
	var unit string
	var size float64
	switch {
	case bytes >= GB:
		unit = "GB"
		size = float64(bytes) / GB
	case bytes >= MB:
		unit = "MB"
		size = float64(bytes) / MB
	case bytes >= KB:
		unit = "KB"
		size = float64(bytes) / KB
	default:
		return fmt.Sprintf("%.2fB", size)
	}
	return fmt.Sprintf("%.2f%s", size, unit)
}

func DownloadPercentage(downloaded, size int64) string {
	return fmt.Sprintf("%.1f", float64(downloaded)/float64(size)*100)
}
