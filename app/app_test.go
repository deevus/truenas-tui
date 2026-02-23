package app_test

import (
	"context"
	"testing"

	"git.sr.ht/~rockorager/vaxis"
	"git.sr.ht/~rockorager/vaxis/vxfw"
	"github.com/deevus/truenas-go"
	"github.com/deevus/truenas-tui/app"
	"github.com/deevus/truenas-tui/internal"
)

func testDrawContext(w, h uint16) vxfw.DrawContext {
	return vxfw.DrawContext{
		Max: vxfw.Size{Width: w, Height: h},
		Min: vxfw.Size{},
		Characters: func(s string) []vaxis.Character {
			chars := make([]vaxis.Character, 0, len(s))
			for _, r := range s {
				chars = append(chars, vaxis.Character{Grapheme: string(r), Width: 1})
			}
			return chars
		},
	}
}

func newTestServices() *internal.Services {
	return internal.NewServices(
		&truenas.MockDatasetService{},
		&truenas.MockSnapshotService{},
	)
}

func newTestServicesWithData() *internal.Services {
	return internal.NewServices(
		&truenas.MockDatasetService{
			ListPoolsFunc: func(ctx context.Context) ([]truenas.Pool, error) {
				return []truenas.Pool{
					{ID: 1, Name: "tank", Status: "ONLINE", Size: 1099511627776},
				}, nil
			},
			ListDatasetsFunc: func(ctx context.Context) ([]truenas.Dataset, error) {
				return []truenas.Dataset{
					{ID: "tank/data", Name: "data", Pool: "tank", Compression: "lz4"},
				}, nil
			},
		},
		&truenas.MockSnapshotService{
			ListFunc: func(ctx context.Context) ([]truenas.Snapshot, error) {
				return []truenas.Snapshot{
					{ID: "tank/data@snap1", Dataset: "tank/data", SnapshotName: "snap1"},
				}, nil
			},
		},
	)
}

func TestApp_New(t *testing.T) {
	a := app.New(newTestServices(), "test-server")
	if a == nil {
		t.Fatal("expected non-nil app")
	}
}

func TestApp_ActiveTab(t *testing.T) {
	a := app.New(newTestServices(), "test-server")
	if a.ActiveTab() != 0 {
		t.Errorf("expected initial tab 0, got %d", a.ActiveTab())
	}
}

func TestApp_SetTab(t *testing.T) {
	a := app.New(newTestServices(), "test-server")
	a.SetTab(1)
	if a.ActiveTab() != 1 {
		t.Errorf("expected tab 1, got %d", a.ActiveTab())
	}
	a.SetTab(2)
	if a.ActiveTab() != 2 {
		t.Errorf("expected tab 2, got %d", a.ActiveTab())
	}
}

func TestApp_ServerName(t *testing.T) {
	a := app.New(newTestServices(), "home")
	if a.ServerName() != "home" {
		t.Errorf("expected server name home, got %s", a.ServerName())
	}
}

func TestApp_LoadActiveView_Tab0_Pools(t *testing.T) {
	svc := newTestServicesWithData()
	a := app.New(svc, "test-server")
	a.SetTab(0)

	err := a.LoadActiveView(context.Background())
	if err != nil {
		t.Fatalf("unexpected error loading pools: %v", err)
	}
}

func TestApp_LoadActiveView_Tab1_Datasets(t *testing.T) {
	svc := newTestServicesWithData()
	a := app.New(svc, "test-server")
	a.SetTab(1)

	err := a.LoadActiveView(context.Background())
	if err != nil {
		t.Fatalf("unexpected error loading datasets: %v", err)
	}
}

func TestApp_LoadActiveView_Tab2_Snapshots(t *testing.T) {
	svc := newTestServicesWithData()
	a := app.New(svc, "test-server")
	a.SetTab(2)

	err := a.LoadActiveView(context.Background())
	if err != nil {
		t.Fatalf("unexpected error loading snapshots: %v", err)
	}
}

func TestApp_LoadActiveView_Error_Propagation(t *testing.T) {
	svc := internal.NewServices(
		&truenas.MockDatasetService{
			ListPoolsFunc: func(ctx context.Context) ([]truenas.Pool, error) {
				return nil, context.DeadlineExceeded
			},
		},
		&truenas.MockSnapshotService{},
	)
	a := app.New(svc, "test-server")
	a.SetTab(0)

	err := a.LoadActiveView(context.Background())
	if err == nil {
		t.Fatal("expected error to propagate from pools Load")
	}
}

func TestApp_LoadActiveView_Error_Datasets(t *testing.T) {
	svc := internal.NewServices(
		&truenas.MockDatasetService{
			ListDatasetsFunc: func(ctx context.Context) ([]truenas.Dataset, error) {
				return nil, context.DeadlineExceeded
			},
		},
		&truenas.MockSnapshotService{},
	)
	a := app.New(svc, "test-server")
	a.SetTab(1)

	err := a.LoadActiveView(context.Background())
	if err == nil {
		t.Fatal("expected error to propagate from datasets Load")
	}
}

func TestApp_LoadActiveView_Error_Snapshots(t *testing.T) {
	svc := internal.NewServices(
		&truenas.MockDatasetService{},
		&truenas.MockSnapshotService{
			ListFunc: func(ctx context.Context) ([]truenas.Snapshot, error) {
				return nil, context.DeadlineExceeded
			},
		},
	)
	a := app.New(svc, "test-server")
	a.SetTab(2)

	err := a.LoadActiveView(context.Background())
	if err == nil {
		t.Fatal("expected error to propagate from snapshots Load")
	}
}

func TestApp_Draw(t *testing.T) {
	svc := newTestServicesWithData()
	a := app.New(svc, "test-server")
	_ = a.LoadActiveView(context.Background())

	ctx := testDrawContext(80, 24)
	s, err := a.Draw(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if s.Size.Width != 80 {
		t.Errorf("expected surface width=80, got %d", s.Size.Width)
	}
	if s.Size.Height != 24 {
		t.Errorf("expected surface height=24, got %d", s.Size.Height)
	}
}

func TestApp_Draw_AllTabs(t *testing.T) {
	svc := newTestServicesWithData()
	a := app.New(svc, "test-server")

	for tab := 0; tab < 3; tab++ {
		a.SetTab(tab)
		_ = a.LoadActiveView(context.Background())

		ctx := testDrawContext(80, 24)
		_, err := a.Draw(ctx)
		if err != nil {
			t.Fatalf("unexpected error drawing tab %d: %v", tab, err)
		}
	}
}

func TestApp_CaptureEvent_Quit(t *testing.T) {
	a := app.New(newTestServices(), "test-server")

	cmd, err := a.CaptureEvent(vaxis.Key{Keycode: 'q'})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, ok := cmd.(vxfw.QuitCmd); !ok {
		t.Errorf("expected QuitCmd, got %T", cmd)
	}
}

func TestApp_CaptureEvent_NumberKeys(t *testing.T) {
	a := app.New(newTestServices(), "test-server")

	tests := []struct {
		key      rune
		expected int
	}{
		{'1', 0},
		{'2', 1},
		{'3', 2},
	}

	for _, tc := range tests {
		cmd, err := a.CaptureEvent(vaxis.Key{Keycode: tc.key})
		if err != nil {
			t.Fatalf("unexpected error for key '%c': %v", tc.key, err)
		}
		if cmd == nil {
			t.Fatalf("expected non-nil command for key '%c'", tc.key)
		}
		if a.ActiveTab() != tc.expected {
			t.Errorf("key '%c': expected tab %d, got %d", tc.key, tc.expected, a.ActiveTab())
		}
	}
}

func TestApp_CaptureEvent_Tab(t *testing.T) {
	a := app.New(newTestServices(), "test-server")

	// Tab cycles forward
	cmd, err := a.CaptureEvent(vaxis.Key{Keycode: vaxis.KeyTab})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cmd == nil {
		t.Fatal("expected non-nil command for Tab key")
	}
	if a.ActiveTab() != 1 {
		t.Errorf("expected tab 1 after Tab, got %d", a.ActiveTab())
	}
}

func TestApp_CaptureEvent_ShiftTab(t *testing.T) {
	a := app.New(newTestServices(), "test-server")

	// Shift-Tab cycles backward (wraps from 0 to 2)
	cmd, err := a.CaptureEvent(vaxis.Key{Keycode: vaxis.KeyTab, Modifiers: vaxis.ModShift})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cmd == nil {
		t.Fatal("expected non-nil command for Shift+Tab")
	}
	if a.ActiveTab() != 2 {
		t.Errorf("expected tab 2 after Shift+Tab, got %d", a.ActiveTab())
	}
}

func TestApp_CaptureEvent_UnhandledKey(t *testing.T) {
	a := app.New(newTestServices(), "test-server")

	cmd, err := a.CaptureEvent(vaxis.Key{Keycode: 'x'})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cmd != nil {
		t.Errorf("expected nil command for unhandled key, got %T", cmd)
	}
}

func TestApp_CaptureEvent_NonKeyEvent(t *testing.T) {
	a := app.New(newTestServices(), "test-server")

	// Pass a non-key event (e.g., a Redraw event)
	cmd, err := a.CaptureEvent(vaxis.Redraw{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cmd != nil {
		t.Errorf("expected nil command for non-key event, got %T", cmd)
	}
}

func TestApp_HandleEvent(t *testing.T) {
	svc := newTestServicesWithData()
	a := app.New(svc, "test-server")
	_ = a.LoadActiveView(context.Background())

	// HandleEvent delegates to the active view
	cmd, err := a.HandleEvent(vaxis.Key{Keycode: 'j'}, vxfw.EventPhase(0))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	_ = cmd
}

func TestApp_HandleEvent_AllTabs(t *testing.T) {
	svc := newTestServicesWithData()
	a := app.New(svc, "test-server")

	for tab := 0; tab < 3; tab++ {
		a.SetTab(tab)
		_ = a.LoadActiveView(context.Background())

		cmd, err := a.HandleEvent(vaxis.Key{Keycode: 'j'}, vxfw.EventPhase(0))
		if err != nil {
			t.Fatalf("unexpected error on tab %d: %v", tab, err)
		}
		_ = cmd
	}
}
