package app_test

import (
	"context"
	"sync"
	"testing"
	"time"

	"git.sr.ht/~rockorager/vaxis"
	"git.sr.ht/~rockorager/vaxis/vxfw"
	"github.com/deevus/truenas-go"
	"github.com/deevus/truenas-tui/app"
	"github.com/deevus/truenas-tui/internal"
	"github.com/deevus/truenas-tui/views"
)

const testStaleTTL = 30 * time.Second

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
	a := app.New(newTestServices(), "test-server", testStaleTTL)
	if a == nil {
		t.Fatal("expected non-nil app")
	}
}

func TestApp_ActiveTab(t *testing.T) {
	a := app.New(newTestServices(), "test-server", testStaleTTL)
	if a.ActiveTab() != 0 {
		t.Errorf("expected initial tab 0, got %d", a.ActiveTab())
	}
}

func TestApp_SetTab(t *testing.T) {
	a := app.New(newTestServices(), "test-server", testStaleTTL)
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
	a := app.New(newTestServices(), "home", testStaleTTL)
	if a.ServerName() != "home" {
		t.Errorf("expected server name home, got %s", a.ServerName())
	}
}

func TestApp_LoadActiveView_Tab0_Pools(t *testing.T) {
	svc := newTestServicesWithData()
	a := app.New(svc, "test-server", testStaleTTL)
	a.SetTab(0)

	err := a.LoadActiveView(context.Background())
	if err != nil {
		t.Fatalf("unexpected error loading pools: %v", err)
	}
}

func TestApp_LoadActiveView_Tab1_Datasets(t *testing.T) {
	svc := newTestServicesWithData()
	a := app.New(svc, "test-server", testStaleTTL)
	a.SetTab(1)

	err := a.LoadActiveView(context.Background())
	if err != nil {
		t.Fatalf("unexpected error loading datasets: %v", err)
	}
}

func TestApp_LoadActiveView_Tab2_Snapshots(t *testing.T) {
	svc := newTestServicesWithData()
	a := app.New(svc, "test-server", testStaleTTL)
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
	a := app.New(svc, "test-server", testStaleTTL)
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
	a := app.New(svc, "test-server", testStaleTTL)
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
	a := app.New(svc, "test-server", testStaleTTL)
	a.SetTab(2)

	err := a.LoadActiveView(context.Background())
	if err == nil {
		t.Fatal("expected error to propagate from snapshots Load")
	}
}

func TestApp_Draw(t *testing.T) {
	svc := newTestServicesWithData()
	a := app.New(svc, "test-server", testStaleTTL)
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
	a := app.New(svc, "test-server", testStaleTTL)

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

func TestApp_Draw_BeforeLoad(t *testing.T) {
	a := app.New(newTestServices(), "test-server", testStaleTTL)

	ctx := testDrawContext(80, 24)
	_, err := a.Draw(ctx)
	if err != nil {
		t.Fatalf("unexpected error drawing before load: %v", err)
	}
}

func TestApp_CaptureEvent_Quit(t *testing.T) {
	a := app.New(newTestServices(), "test-server", testStaleTTL)

	cmd, err := a.CaptureEvent(vaxis.Key{Keycode: 'q'})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, ok := cmd.(vxfw.QuitCmd); !ok {
		t.Errorf("expected QuitCmd, got %T", cmd)
	}
}

func TestApp_CaptureEvent_NumberKeys(t *testing.T) {
	a := app.New(newTestServices(), "test-server", testStaleTTL)

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
	a := app.New(newTestServices(), "test-server", testStaleTTL)

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
	a := app.New(newTestServices(), "test-server", testStaleTTL)

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
	a := app.New(newTestServices(), "test-server", testStaleTTL)

	cmd, err := a.CaptureEvent(vaxis.Key{Keycode: 'x'})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cmd != nil {
		t.Errorf("expected nil command for unhandled key, got %T", cmd)
	}
}

func TestApp_CaptureEvent_NonKeyEvent(t *testing.T) {
	a := app.New(newTestServices(), "test-server", testStaleTTL)

	// Pass a non-key event (e.g., a Redraw event)
	cmd, err := a.CaptureEvent(vaxis.Redraw{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cmd != nil {
		t.Errorf("expected nil command for non-key event, got %T", cmd)
	}
}

func TestApp_CaptureEvent_Refresh(t *testing.T) {
	svc := newTestServicesWithData()
	a := app.New(svc, "test-server", testStaleTTL)

	cmd, err := a.CaptureEvent(vaxis.Key{Keycode: 'r'})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cmd == nil {
		t.Fatal("expected non-nil command for 'r' key")
	}
}

func TestApp_HandleEvent(t *testing.T) {
	svc := newTestServicesWithData()
	a := app.New(svc, "test-server", testStaleTTL)
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
	a := app.New(svc, "test-server", testStaleTTL)

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

func TestApp_HandleEvent_ViewLoaded(t *testing.T) {
	a := app.New(newTestServices(), "test-server", testStaleTTL)

	cmd, err := a.HandleEvent(views.ViewLoaded{Tab: 0, Err: nil}, vxfw.EventPhase(0))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, ok := cmd.(vxfw.RedrawCmd); !ok {
		t.Errorf("expected RedrawCmd, got %T", cmd)
	}
}

func TestApp_HandleEvent_ViewLoaded_WithError(t *testing.T) {
	a := app.New(newTestServices(), "test-server", testStaleTTL)

	cmd, err := a.HandleEvent(views.ViewLoaded{Tab: 1, Err: context.DeadlineExceeded}, vxfw.EventPhase(0))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, ok := cmd.(vxfw.RedrawCmd); !ok {
		t.Errorf("expected RedrawCmd even on load error, got %T", cmd)
	}
}

func TestApp_LoadAll(t *testing.T) {
	svc := newTestServicesWithData()
	a := app.New(svc, "test-server", testStaleTTL)

	var mu sync.Mutex
	var events []views.ViewLoaded
	done := make(chan struct{}, 3)

	a.SetPostEvent(func(ev vaxis.Event) {
		if vl, ok := ev.(views.ViewLoaded); ok {
			mu.Lock()
			events = append(events, vl)
			mu.Unlock()
			done <- struct{}{}
		}
	})

	a.LoadAll(context.Background())

	// Wait for all 3 goroutines to complete
	for i := 0; i < 3; i++ {
		select {
		case <-done:
		case <-time.After(5 * time.Second):
			t.Fatal("timed out waiting for LoadAll to complete")
		}
	}

	mu.Lock()
	defer mu.Unlock()

	if len(events) != 3 {
		t.Fatalf("expected 3 ViewLoaded events, got %d", len(events))
	}

	tabs := map[int]bool{}
	for _, ev := range events {
		if ev.Err != nil {
			t.Errorf("tab %d had unexpected error: %v", ev.Tab, ev.Err)
		}
		tabs[ev.Tab] = true
	}
	for i := 0; i < 3; i++ {
		if !tabs[i] {
			t.Errorf("missing ViewLoaded event for tab %d", i)
		}
	}
}

func TestApp_LoadAll_WithErrors(t *testing.T) {
	svc := internal.NewServices(
		&truenas.MockDatasetService{
			ListPoolsFunc: func(ctx context.Context) ([]truenas.Pool, error) {
				return nil, context.DeadlineExceeded
			},
		},
		&truenas.MockSnapshotService{
			ListFunc: func(ctx context.Context) ([]truenas.Snapshot, error) {
				return nil, context.Canceled
			},
		},
	)
	a := app.New(svc, "test-server", testStaleTTL)

	var mu sync.Mutex
	var events []views.ViewLoaded
	done := make(chan struct{}, 3)

	a.SetPostEvent(func(ev vaxis.Event) {
		if vl, ok := ev.(views.ViewLoaded); ok {
			mu.Lock()
			events = append(events, vl)
			mu.Unlock()
			done <- struct{}{}
		}
	})

	a.LoadAll(context.Background())

	for i := 0; i < 3; i++ {
		select {
		case <-done:
		case <-time.After(5 * time.Second):
			t.Fatal("timed out waiting for LoadAll to complete")
		}
	}

	mu.Lock()
	defer mu.Unlock()

	errCount := 0
	for _, ev := range events {
		if ev.Err != nil {
			errCount++
		}
	}
	if errCount == 0 {
		t.Error("expected at least one error from LoadAll")
	}
}

func TestApp_TabSwitch_RefetchesStale(t *testing.T) {
	callCount := 0
	svc := internal.NewServices(
		&truenas.MockDatasetService{
			ListDatasetsFunc: func(ctx context.Context) ([]truenas.Dataset, error) {
				callCount++
				return []truenas.Dataset{
					{ID: "tank/data", Name: "data", Pool: "tank"},
				}, nil
			},
		},
		&truenas.MockSnapshotService{},
	)

	// Use zero TTL so data is always stale after load
	a := app.New(svc, "test-server", 0)
	a.SetTab(1)
	_ = a.LoadActiveView(context.Background())
	initial := callCount

	// Small delay so time.Since(loadedAt) > 0
	time.Sleep(time.Millisecond)

	// Switch away and back to datasets tab — should trigger refetch
	a.SetTab(0)
	_, _ = a.CaptureEvent(vaxis.Key{Keycode: '2'})

	if callCount <= initial {
		t.Error("expected refetch on stale tab switch")
	}
}

func TestApp_TabSwitch_NoRefetchWhenFresh(t *testing.T) {
	callCount := 0
	svc := internal.NewServices(
		&truenas.MockDatasetService{
			ListDatasetsFunc: func(ctx context.Context) ([]truenas.Dataset, error) {
				callCount++
				return []truenas.Dataset{
					{ID: "tank/data", Name: "data", Pool: "tank"},
				}, nil
			},
		},
		&truenas.MockSnapshotService{},
	)

	// Use large TTL so data stays fresh
	a := app.New(svc, "test-server", time.Hour)
	a.SetTab(1)
	_ = a.LoadActiveView(context.Background())
	afterLoad := callCount

	// Switch away and back — should NOT refetch
	a.SetTab(0)
	_, _ = a.CaptureEvent(vaxis.Key{Keycode: '2'})

	if callCount != afterLoad {
		t.Errorf("expected no refetch when fresh, but ListDatasetsFunc called %d more times", callCount-afterLoad)
	}
}
