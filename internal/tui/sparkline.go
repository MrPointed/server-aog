package tui

import (
	"strings"
)

var sparks = []string{" ", "▂", "▃", "▄", "▅", "▆", "▇", "█"}

func renderSparkline(values []uint64, width int) string {
	if len(values) == 0 {
		return strings.Repeat(" ", width)
	}

	// 1. Find Max to normalize
	var max uint64
	for _, v := range values {
		if v > max {
			max = v
		}
	}

	// Avoid division by zero
	if max == 0 {
		return strings.Repeat(" ", width)
	}

	// 2. Generate string
	var sb strings.Builder
	
	// We want to show the last 'width' values. 
	// If we have fewer, pad with spaces or just show what we have.
	// Actually, the prompt implies the history size matches visual width roughly.
	
	start := 0
	if len(values) > width {
		start = len(values) - width
	}
	
	visible := values[start:]
	
	// Pad if needed (though usually we fill from right)
	if len(visible) < width {
		sb.WriteString(strings.Repeat(" ", width-len(visible)))
	}

	for _, v := range visible {
		// Index in sparks array: (v * (len(sparks)-1)) / max
		idx := int((v * uint64(len(sparks)-1)) / max)
		if idx >= len(sparks) {
			idx = len(sparks) - 1
		}
		sb.WriteString(sparks[idx])
	}

	return sb.String()
}
