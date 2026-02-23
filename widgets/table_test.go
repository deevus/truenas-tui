package widgets_test

import (
	"testing"

	"git.sr.ht/~rockorager/vaxis"
	"github.com/deevus/truenas-tui/widgets"
)

func cellText(s vaxis.Cell) string {
	return s.Character.Grapheme
}

func TestTable_Draw_AlignedColumns(t *testing.T) {
	tbl := &widgets.Table{
		Columns: []widgets.TableColumn{
			{Width: 10},
			{Width: 6, AlignRight: true},
			{Width: 8, AlignRight: true},
		},
		Header: []string{"NAME", "CPU%", "MEM"},
		Rows: [][]string{
			{"tailscale", "0.28%", "142 MB"},
			{"dns", "0.38%", "138 MB"},
		},
		Gap: 2,
	}

	ctx := testDrawContext(40, 10)
	surf, err := tbl.Draw(ctx)
	if err != nil {
		t.Fatalf("Draw: %v", err)
	}

	// Should have 3 rows: header + 2 data
	if surf.Size.Height != 3 {
		t.Fatalf("expected height=3, got %d", surf.Size.Height)
	}

	// Check header row: "NAME" starts at col 0
	if g := cellText(surf.Buffer[0]); g != "N" {
		t.Errorf("header col 0: expected 'N', got %q", g)
	}

	// "CPU%" is right-aligned in width 6 starting at col 12 (10+2 gap).
	// "CPU%" is 4 chars, right-aligned in 6 = 2 offset, so col 14.
	if g := cellText(surf.Buffer[14]); g != "C" {
		t.Errorf("header CPU col 14: expected 'C', got %q", g)
	}

	// Data row 1 starts at row 1 (offset = 1*40 = 40)
	// "tailscale" at col 0
	if g := cellText(surf.Buffer[40]); g != "t" {
		t.Errorf("row1 col 0: expected 't', got %q", g)
	}

	// "0.28%" is 5 chars, right-aligned in 6 = 1 offset, col 13
	if g := cellText(surf.Buffer[40+13]); g != "0" {
		t.Errorf("row1 cpu col 13: expected '0', got %q", g)
	}

	// Row 2: "dns" at col 0
	if g := cellText(surf.Buffer[80]); g != "d" {
		t.Errorf("row2 col 0: expected 'd', got %q", g)
	}

	// Both "0.28%" and "0.38%" should start at the same column (13)
	if g := cellText(surf.Buffer[80+13]); g != "0" {
		t.Errorf("row2 cpu col 13: expected '0', got %q", g)
	}
}

func TestTable_Draw_NoHeader(t *testing.T) {
	tbl := &widgets.Table{
		Columns: []widgets.TableColumn{
			{Width: 8},
			{Width: 6},
		},
		Rows: [][]string{
			{"hello", "world"},
		},
	}

	ctx := testDrawContext(30, 5)
	surf, err := tbl.Draw(ctx)
	if err != nil {
		t.Fatalf("Draw: %v", err)
	}

	if surf.Size.Height != 1 {
		t.Errorf("expected height=1 (no header), got %d", surf.Size.Height)
	}
}

func TestTable_Draw_TruncatesLongText(t *testing.T) {
	tbl := &widgets.Table{
		Columns: []widgets.TableColumn{
			{Width: 4},
		},
		Rows: [][]string{
			{"toolongname"},
		},
	}

	ctx := testDrawContext(20, 5)
	surf, err := tbl.Draw(ctx)
	if err != nil {
		t.Fatalf("Draw: %v", err)
	}

	// Should only write 4 chars
	if g := cellText(surf.Buffer[0]); g != "t" {
		t.Errorf("col 0: expected 't', got %q", g)
	}
	if g := cellText(surf.Buffer[3]); g != "l" {
		t.Errorf("col 3: expected 'l', got %q", g)
	}
	// Col 4 should be empty (beyond column width)
	if g := cellText(surf.Buffer[4]); g != "" {
		t.Errorf("col 4: expected empty, got %q", g)
	}
}
