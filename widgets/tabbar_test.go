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
