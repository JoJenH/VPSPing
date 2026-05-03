package stats

import (
	"fmt"
	"math"
	"strings"
	"time"
)

type Chart struct {
	width  int
	height int
}

func NewChart(width, height int) *Chart {
	return &Chart{
		width:  width,
		height: height,
	}
}

func (c *Chart) DrawLineChart(points []LatencyPoint, title string) string {
	if len(points) == 0 {
		return "No data available"
	}

	minVal, maxVal := c.getMinMax(points)
	if maxVal == minVal {
		maxVal = minVal + 1
	}

	chart := make([][]rune, c.height)
	for i := range chart {
		chart[i] = make([]rune, c.width)
		for j := range chart[i] {
			chart[i][j] = ' '
		}
	}

	xStep := float64(c.width-1) / float64(len(points)-1)
	for i, point := range points {
		x := int(float64(i) * xStep)
		if x >= c.width {
			x = c.width - 1
		}

		y := int((maxVal - point.Value) / (maxVal - minVal) * float64(c.height-1))
		if y < 0 {
			y = 0
		}
		if y >= c.height {
			y = c.height - 1
		}

		chart[y][x] = '●'
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("\n%s\n", title))
	sb.WriteString(fmt.Sprintf("  %.0fms ┤\n", maxVal))

	for i := 0; i < c.height; i++ {
		if i == 0 || i == c.height-1 || i == c.height/2 {
			value := maxVal - (maxVal-minVal)*float64(i)/float64(c.height-1)
			sb.WriteString(fmt.Sprintf("  %5.0fms ┤", value))
		} else {
			sb.WriteString("         │")
		}

		for j := 0; j < c.width; j++ {
			sb.WriteRune(chart[i][j])
		}
		sb.WriteString("\n")
	}

	sb.WriteString("         └")
	for i := 0; i < c.width; i++ {
		sb.WriteRune('─')
	}
	sb.WriteString("\n")

	if len(points) > 1 {
		startTime := points[0].Timestamp.Format("15:04")
		endTime := points[len(points)-1].Timestamp.Format("15:04")
		sb.WriteString(fmt.Sprintf("          %s", startTime))
		for i := 0; i < c.width-15; i++ {
			sb.WriteRune(' ')
		}
		sb.WriteString(fmt.Sprintf("%s\n", endTime))
	}

	return sb.String()
}

func (c *Chart) DrawMultiLineChart(dataSets map[string][]LatencyPoint, title string) string {
	if len(dataSets) == 0 {
		return "No data available"
	}

	var allPoints []LatencyPoint
	for _, points := range dataSets {
		allPoints = append(allPoints, points...)
	}

	if len(allPoints) == 0 {
		return "No data available"
	}

	minVal, maxVal := c.getMinMax(allPoints)
	if maxVal == minVal {
		maxVal = minVal + 1
	}

	chart := make([][]rune, c.height)
	for i := range chart {
		chart[i] = make([]rune, c.width)
		for j := range chart[i] {
			chart[i][j] = ' '
		}
	}

	symbols := []rune{'●', '◆', '▲', '■', '★'}
	colorIdx := 0

	for _, points := range dataSets {
		if len(points) == 0 {
			continue
		}

		symbol := symbols[colorIdx%len(symbols)]
		colorIdx++

		xStep := float64(c.width-1) / float64(len(points)-1)
		for i, point := range points {
			x := int(float64(i) * xStep)
			if x >= c.width {
				x = c.width - 1
			}

			y := int((maxVal - point.Value) / (maxVal - minVal) * float64(c.height-1))
			if y < 0 {
				y = 0
			}
			if y >= c.height {
				y = c.height - 1
			}

			chart[y][x] = symbol
		}
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("\n%s\n", title))
	sb.WriteString(fmt.Sprintf("  %.0fms ┤\n", maxVal))

	for i := 0; i < c.height; i++ {
		if i == 0 || i == c.height-1 || i == c.height/2 {
			value := maxVal - (maxVal-minVal)*float64(i)/float64(c.height-1)
			sb.WriteString(fmt.Sprintf("  %5.0fms ┤", value))
		} else {
			sb.WriteString("         │")
		}

		for j := 0; j < c.width; j++ {
			sb.WriteRune(chart[i][j])
		}
		sb.WriteString("\n")
	}

	sb.WriteString("         └")
	for i := 0; i < c.width; i++ {
		sb.WriteRune('─')
	}
	sb.WriteString("\n")

	sb.WriteString("\nLegend:\n")
	colorIdx = 0
	for name := range dataSets {
		symbol := symbols[colorIdx%len(symbols)]
		sb.WriteString(fmt.Sprintf("  %c %s\n", symbol, name))
		colorIdx++
	}

	return sb.String()
}

func (c *Chart) getMinMax(points []LatencyPoint) (float64, float64) {
	if len(points) == 0 {
		return 0, 100
	}

	minVal := math.MaxFloat64
	maxVal := 0.0

	for _, p := range points {
		if p.Value < minVal {
			minVal = p.Value
		}
		if p.Value > maxVal {
			maxVal = p.Value
		}
	}

	padding := (maxVal - minVal) * 0.1
	if padding == 0 {
		padding = 10
	}

	return minVal - padding, maxVal + padding
}

func FormatDuration(d time.Duration) string {
	if d < time.Minute {
		return fmt.Sprintf("%ds", int(d.Seconds()))
	} else if d < time.Hour {
		return fmt.Sprintf("%dm", int(d.Minutes()))
	} else if d < 24*time.Hour {
		return fmt.Sprintf("%dh", int(d.Hours()))
	}
	return fmt.Sprintf("%dd", int(d.Hours()/24))
}
