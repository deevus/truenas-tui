package widgets_test

import (
	"testing"

	"github.com/deevus/truenas-tui/widgets"
)

func TestSparkline_New(t *testing.T) {
	sl := widgets.NewSparkline(60)
	if sl.Count() != 0 {
		t.Errorf("expected count=0, got %d", sl.Count())
	}
}

func TestSparkline_Push(t *testing.T) {
	sl := widgets.NewSparkline(5)
	sl.Push(10)
	sl.Push(20)
	sl.Push(30)
	if sl.Count() != 3 {
		t.Errorf("expected count=3, got %d", sl.Count())
	}
}

func TestSparkline_Push_WrapsAround(t *testing.T) {
	sl := widgets.NewSparkline(3)
	sl.Push(10)
	sl.Push(20)
	sl.Push(30)
	sl.Push(40) // overwrites 10
	if sl.Count() != 3 {
		t.Errorf("expected count=3 after overflow, got %d", sl.Count())
	}
}

func TestSparkline_Draw_Empty(t *testing.T) {
	sl := widgets.NewSparkline(10)
	ctx := testDrawContext(20, 1)
	s, err := sl.Draw(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if s.Size.Height != 1 {
		t.Errorf("expected height=1, got %d", s.Size.Height)
	}
}

func TestSparkline_Draw_WithData(t *testing.T) {
	sl := widgets.NewSparkline(60)
	for i := 0; i < 10; i++ {
		sl.Push(float64(i * 10))
	}

	ctx := testDrawContext(20, 1)
	_, err := sl.Draw(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestSparkline_Draw_MoreDataThanWidth(t *testing.T) {
	sl := widgets.NewSparkline(60)
	for i := 0; i < 60; i++ {
		sl.Push(float64(i))
	}

	ctx := testDrawContext(10, 1)
	_, err := sl.Draw(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestSparkline_Draw_FlatLine(t *testing.T) {
	sl := widgets.NewSparkline(10)
	for i := 0; i < 5; i++ {
		sl.Push(50.0)
	}

	ctx := testDrawContext(20, 1)
	_, err := sl.Draw(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestSparkline_Draw_FlatZero(t *testing.T) {
	sl := widgets.NewSparkline(10)
	for i := 0; i < 5; i++ {
		sl.Push(0.0)
	}

	ctx := testDrawContext(20, 1)
	_, err := sl.Draw(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
