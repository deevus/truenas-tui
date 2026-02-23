package widgets_test

import (
	"testing"

	"github.com/deevus/truenas-tui/widgets"
)

func TestBarGauge_Draw(t *testing.T) {
	bg := &widgets.BarGauge{
		Label:    "CPU",
		Value:    42.5,
		Suffix:   "65°C",
		BarWidth: 20,
	}

	ctx := testDrawContext(80, 1)
	s, err := bg.Draw(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if s.Size.Height != 1 {
		t.Errorf("expected height=1, got %d", s.Size.Height)
	}
}

func TestBarGauge_Draw_Zero(t *testing.T) {
	bg := &widgets.BarGauge{
		Label:    "MEM",
		Value:    0,
		BarWidth: 20,
	}

	ctx := testDrawContext(80, 1)
	_, err := bg.Draw(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestBarGauge_Draw_Full(t *testing.T) {
	bg := &widgets.BarGauge{
		Label:    "DISK",
		Value:    100,
		BarWidth: 20,
	}

	ctx := testDrawContext(80, 1)
	_, err := bg.Draw(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestBarGauge_Draw_ClampNegative(t *testing.T) {
	bg := &widgets.BarGauge{
		Label:    "ARC",
		Value:    -10,
		BarWidth: 20,
	}

	ctx := testDrawContext(80, 1)
	_, err := bg.Draw(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestBarGauge_Draw_ClampOver100(t *testing.T) {
	bg := &widgets.BarGauge{
		Label:    "ARC",
		Value:    120,
		BarWidth: 20,
	}

	ctx := testDrawContext(80, 1)
	_, err := bg.Draw(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestBarGauge_Draw_WithSuffix(t *testing.T) {
	bg := &widgets.BarGauge{
		Label:    "MEM",
		Value:    81.3,
		Suffix:   "13.1/16.0 GiB",
		BarWidth: 20,
	}

	ctx := testDrawContext(80, 1)
	_, err := bg.Draw(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestBarGauge_Draw_NoSuffix(t *testing.T) {
	bg := &widgets.BarGauge{
		Label:    "ARC",
		Value:    25.0,
		BarWidth: 20,
	}

	ctx := testDrawContext(80, 1)
	_, err := bg.Draw(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
