package widgets_test

import (
	"testing"

	"github.com/deevus/truenas-tui/widgets"
)

func TestTabBar_Labels(t *testing.T) {
	tb := widgets.NewTabBar([]string{"Pools", "Datasets", "Snapshots"})
	if tb.Active() != 0 {
		t.Errorf("expected initial active=0, got %d", tb.Active())
	}
}

func TestTabBar_Next(t *testing.T) {
	tb := widgets.NewTabBar([]string{"A", "B", "C"})
	tb.Next()
	if tb.Active() != 1 {
		t.Errorf("expected active=1, got %d", tb.Active())
	}
	tb.Next()
	if tb.Active() != 2 {
		t.Errorf("expected active=2, got %d", tb.Active())
	}
	// Wraps around
	tb.Next()
	if tb.Active() != 0 {
		t.Errorf("expected active=0 after wrap, got %d", tb.Active())
	}
}

func TestTabBar_Prev(t *testing.T) {
	tb := widgets.NewTabBar([]string{"A", "B", "C"})
	// Wraps backward
	tb.Prev()
	if tb.Active() != 2 {
		t.Errorf("expected active=2 after backward wrap, got %d", tb.Active())
	}
	tb.Prev()
	if tb.Active() != 1 {
		t.Errorf("expected active=1, got %d", tb.Active())
	}
}

func TestTabBar_SetActive(t *testing.T) {
	tb := widgets.NewTabBar([]string{"A", "B", "C"})
	tb.SetActive(2)
	if tb.Active() != 2 {
		t.Errorf("expected active=2, got %d", tb.Active())
	}
	// Out of bounds is clamped
	tb.SetActive(5)
	if tb.Active() != 2 {
		t.Errorf("expected active=2 (clamped), got %d", tb.Active())
	}
	tb.SetActive(-1)
	if tb.Active() != 2 {
		t.Errorf("expected active=2 (clamped negative), got %d", tb.Active())
	}
}

func TestTabBar_Draw(t *testing.T) {
	tb := widgets.NewTabBar([]string{"Pools", "Datasets", "Snapshots"})
	ctx := testDrawContext(80, 1)

	s, err := tb.Draw(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if s.Size.Height != 1 {
		t.Errorf("expected surface height=1, got %d", s.Size.Height)
	}
	if s.Size.Width != 80 {
		t.Errorf("expected surface width=80, got %d", s.Size.Width)
	}
}

func TestTabBar_Draw_ActiveTabChanges(t *testing.T) {
	tb := widgets.NewTabBar([]string{"A", "B", "C"})
	ctx := testDrawContext(40, 1)

	// Draw with tab 0 active
	s1, err := tb.Draw(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if s1.Size.Height != 1 {
		t.Errorf("expected height=1, got %d", s1.Size.Height)
	}

	// Draw with tab 1 active
	tb.SetActive(1)
	s2, err := tb.Draw(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if s2.Size.Height != 1 {
		t.Errorf("expected height=1, got %d", s2.Size.Height)
	}

	// Draw with tab 2 active
	tb.SetActive(2)
	s3, err := tb.Draw(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if s3.Size.Height != 1 {
		t.Errorf("expected height=1, got %d", s3.Size.Height)
	}
}
