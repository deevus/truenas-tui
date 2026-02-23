package widgets

import (
	"math"

	"git.sr.ht/~rockorager/vaxis"
	"git.sr.ht/~rockorager/vaxis/vxfw"
)

// Block characters for sparkline rendering (8 levels).
var sparkBlocks = [8]rune{'▁', '▂', '▃', '▄', '▅', '▆', '▇', '█'}

// Sparkline renders a 1-row graph of recent values using block characters.
type Sparkline struct {
	values []float64
	head   int
	count  int
}

// NewSparkline creates a Sparkline with the given ring buffer capacity.
func NewSparkline(capacity int) *Sparkline {
	return &Sparkline{
		values: make([]float64, capacity),
	}
}

// Push adds a value to the ring buffer.
func (sl *Sparkline) Push(v float64) {
	sl.values[sl.head] = v
	sl.head = (sl.head + 1) % len(sl.values)
	if sl.count < len(sl.values) {
		sl.count++
	}
}

// Count returns the number of values currently stored.
func (sl *Sparkline) Count() int {
	return sl.count
}

// ordered returns the stored values in chronological order.
func (sl *Sparkline) ordered() []float64 {
	if sl.count == 0 {
		return nil
	}
	out := make([]float64, sl.count)
	start := (sl.head - sl.count + len(sl.values)) % len(sl.values)
	for i := 0; i < sl.count; i++ {
		out[i] = sl.values[(start+i)%len(sl.values)]
	}
	return out
}

// Draw renders the sparkline as a single row.
func (sl *Sparkline) Draw(ctx vxfw.DrawContext) (vxfw.Surface, error) {
	s := vxfw.NewSurface(ctx.Max.Width, 1, sl)

	vals := sl.ordered()
	if len(vals) == 0 {
		return s, nil
	}

	// Limit to available width
	width := int(ctx.Max.Width)
	if len(vals) > width {
		vals = vals[len(vals)-width:]
	}

	// Find min/max for scaling
	minV, maxV := vals[0], vals[0]
	for _, v := range vals[1:] {
		if v < minV {
			minV = v
		}
		if v > maxV {
			maxV = v
		}
	}

	for i, v := range vals {
		level := 0
		if maxV > minV {
			level = int(math.Round((v - minV) / (maxV - minV) * 7))
			if level > 7 {
				level = 7
			}
		} else if maxV > 0 {
			level = 4 // flat non-zero line
		}

		ch := sparkBlocks[level]
		for _, c := range ctx.Characters(string(ch)) {
			s.WriteCell(uint16(i), 0, vaxis.Cell{
				Character: c,
				Style:     vaxis.Style{Foreground: vaxis.IndexColor(6)}, // cyan
			})
		}
	}

	return s, nil
}
