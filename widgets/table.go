package widgets

import (
	"git.sr.ht/~rockorager/vaxis"
	"git.sr.ht/~rockorager/vaxis/vxfw"
)

// TableColumn defines a column in a Table.
type TableColumn struct {
	Width      int         // fixed character width
	AlignRight bool        // right-align text within the column
	Style      vaxis.Style // applied to all cells in this column
}

// Table renders rows of text with fixed-width columns using WriteCell.
// Each row is a []string matching the Columns slice.
type Table struct {
	Columns []TableColumn
	Rows    [][]string
	Header  []string // optional header row rendered with AttrDim
	Gap     int      // spaces between columns (default 1)
}

// writeText writes s into surf at (col, row) within maxWidth, returning
// the number of columns consumed. If right-aligned, text is padded on the left.
func writeText(surf *vxfw.Surface, col, row uint16, maxWidth int, s string, style vaxis.Style, alignRight bool) {
	chars := vaxis.Characters(s)

	// Calculate display width
	displayWidth := 0
	for _, ch := range chars {
		displayWidth += ch.Width
	}

	// Determine starting offset for right alignment
	offset := 0
	if alignRight && displayWidth < maxWidth {
		offset = maxWidth - displayWidth
	}

	pos := offset
	for _, ch := range chars {
		if pos+ch.Width > maxWidth {
			break
		}
		surf.WriteCell(col+uint16(pos), row, vaxis.Cell{
			Character: ch,
			Style:     style,
		})
		pos += ch.Width
	}
}

// Draw renders the table header (if set) and all rows.
func (t *Table) Draw(ctx vxfw.DrawContext) (vxfw.Surface, error) {
	gap := t.Gap
	if gap == 0 {
		gap = 1
	}

	totalRows := len(t.Rows)
	if t.Header != nil {
		totalRows++
	}

	height := uint16(totalRows)
	if height > ctx.Max.Height {
		height = ctx.Max.Height
	}

	s := vxfw.NewSurface(ctx.Max.Width, height, t)
	row := uint16(0)

	// Header
	if t.Header != nil && row < height {
		col := uint16(0)
		for i, c := range t.Columns {
			if int(col) >= int(ctx.Max.Width) {
				break
			}
			text := ""
			if i < len(t.Header) {
				text = t.Header[i]
			}
			style := vaxis.Style{Attribute: vaxis.AttrDim}
			writeText(&s, col, row, c.Width, text, style, c.AlignRight)
			col += uint16(c.Width + gap)
		}
		row++
	}

	// Data rows
	for _, cells := range t.Rows {
		if row >= height {
			break
		}
		col := uint16(0)
		for i, c := range t.Columns {
			if int(col) >= int(ctx.Max.Width) {
				break
			}
			text := ""
			if i < len(cells) {
				text = cells[i]
			}
			style := c.Style
			writeText(&s, col, row, c.Width, text, style, c.AlignRight)
			col += uint16(c.Width + gap)
		}
		row++
	}

	return s, nil
}
