package widgets

import (
	"fmt"

	"git.sr.ht/~rockorager/vaxis"
	"git.sr.ht/~rockorager/vaxis/vxfw"
)

// BarGauge is a horizontal bar gauge widget.
//
//	CPU  [████████░░░░░░░░░░░░]  42.5%  65°C
type BarGauge struct {
	Label    string  // 4-char left column, e.g. "CPU", "MEM"
	Value    float64 // 0.0–100.0
	Suffix   string  // text after %, e.g. "65°C" or "13.1/16.0 GiB"
	BarWidth int     // character width of the [████░░░░] portion (excluding brackets)
}

const (
	barFilled = '█' // U+2588
	barEmpty  = '░' // U+2591
)

// barColor returns the appropriate color for the given percentage.
func barColor(pct float64) vaxis.Color {
	switch {
	case pct >= 85:
		return vaxis.IndexColor(1) // red
	case pct >= 60:
		return vaxis.IndexColor(3) // yellow
	default:
		return vaxis.IndexColor(2) // green
	}
}

// Draw renders the bar gauge as a single row.
func (bg *BarGauge) Draw(ctx vxfw.DrawContext) (vxfw.Surface, error) {
	s := vxfw.NewSurface(ctx.Max.Width, 1, bg)

	col := uint16(0)

	// Label (left-padded to 4 chars)
	label := fmt.Sprintf("%-4s ", bg.Label)
	for _, ch := range ctx.Characters(label) {
		s.WriteCell(col, 0, vaxis.Cell{
			Character: ch,
			Style:     vaxis.Style{Attribute: vaxis.AttrBold},
		})
		col += uint16(ch.Width)
	}

	// Opening bracket
	for _, ch := range ctx.Characters("[") {
		s.WriteCell(col, 0, vaxis.Cell{Character: ch})
		col += uint16(ch.Width)
	}

	// Bar fill
	v := bg.Value
	if v < 0 {
		v = 0
	}
	if v > 100 {
		v = 100
	}
	filled := int(v / 100 * float64(bg.BarWidth))
	color := barColor(v)

	for i := 0; i < bg.BarWidth; i++ {
		ch := barEmpty
		style := vaxis.Style{Foreground: vaxis.IndexColor(8)} // dim for empty
		if i < filled {
			ch = barFilled
			style = vaxis.Style{Foreground: color}
		}
		for _, c := range ctx.Characters(string(ch)) {
			s.WriteCell(col, 0, vaxis.Cell{Character: c, Style: style})
			col += uint16(c.Width)
		}
	}

	// Closing bracket + percentage
	pctStr := fmt.Sprintf("] %5.1f%%", v)
	for _, ch := range ctx.Characters(pctStr) {
		s.WriteCell(col, 0, vaxis.Cell{Character: ch})
		col += uint16(ch.Width)
	}

	// Suffix
	if bg.Suffix != "" {
		suffix := "  " + bg.Suffix
		dimStyle := vaxis.Style{Attribute: vaxis.AttrDim}
		for _, ch := range ctx.Characters(suffix) {
			s.WriteCell(col, 0, vaxis.Cell{Character: ch, Style: dimStyle})
			col += uint16(ch.Width)
		}
	}

	return s, nil
}
