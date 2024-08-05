package main

import "fmt"

const (
	_  = iota
	KB = 1 << (10 * iota)
	MB
	GB
)

func HumanizeSize(bytes float64) string {
	var unit string
	switch {
	case bytes >= GB:
		unit = "GB"
		bytes = bytes / GB
	case bytes >= MB:
		unit = "MB"
		bytes = bytes / MB
	case bytes >= KB:
		unit = "KB"
		bytes = bytes / KB
	default:
		return fmt.Sprintf("%.2fB", bytes)
	}
	return fmt.Sprintf("%.2f%s", bytes, unit)
}
