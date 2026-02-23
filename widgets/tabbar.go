package widgets

import (
	"git.sr.ht/~rockorager/vaxis"
	"git.sr.ht/~rockorager/vaxis/vxfw"
)

// TabBar is a horizontal tab navigation widget.
type TabBar struct {
	labels []string
	active int
}

// NewTabBar creates a TabBar with the given labels. Active defaults to 0.
func NewTabBar(labels []string) *TabBar {
	return &TabBar{labels: labels}
}

// Active returns the currently active tab index.
func (tb *TabBar) Active() int {
	return tb.active
}

// SetActive sets the active tab index. Out-of-range values are ignored.
func (tb *TabBar) SetActive(i int) {
	if i >= 0 && i < len(tb.labels) {
		tb.active = i
	}
}

// Next advances to the next tab, wrapping around.
func (tb *TabBar) Next() {
	tb.active = (tb.active + 1) % len(tb.labels)
}

// Prev moves to the previous tab, wrapping around.
func (tb *TabBar) Prev() {
	tb.active = (tb.active - 1 + len(tb.labels)) % len(tb.labels)
}

// Draw renders the tab bar as a single row: " Pools | Datasets | Snapshots "
// Active tab is rendered with reverse video.
func (tb *TabBar) Draw(ctx vxfw.DrawContext) (vxfw.Surface, error) {
	s := vxfw.NewSurface(ctx.Max.Width, 1, tb)

	col := uint16(0)
	for i, label := range tb.labels {
		if i > 0 {
			// Separator
			sep := " | "
			for _, ch := range ctx.Characters(sep) {
				s.WriteCell(col, 0, vaxis.Cell{Character: ch})
				col += uint16(ch.Width)
			}
		}

		style := vaxis.Style{}
		if i == tb.active {
			style.Attribute |= vaxis.AttrReverse
		}

		text := " " + label + " "
		for _, ch := range ctx.Characters(text) {
			s.WriteCell(col, 0, vaxis.Cell{Character: ch, Style: style})
			col += uint16(ch.Width)
		}
	}

	return s, nil
}
