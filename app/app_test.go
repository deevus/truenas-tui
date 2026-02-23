package app_test

import (
	"context"
	"fmt"
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
		&truenas.MockSystemService{
			GetInfoFunc: func(ctx context.Context) (*truenas.SystemInfo, error) {
				return &truenas.SystemInfo{Hostname: "test", Model: "Test"}, nil
			},
			GetVersionFunc: func(ctx context.Context) (string, error) {
				return "TrueNAS-TEST", nil
			},
		},
		&truenas.MockReportingService{},
		&truenas.MockInterfaceService{
			ListFunc: func(ctx context.Context) ([]truenas.NetworkInterface, error) {
				return nil, nil
			},
		},
		&truenas.MockAppService{
			ListAppsFunc: func(ctx context.Context) ([]truenas.App, error) {
				return nil, nil
			},
		},
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
		&truenas.MockSystemService{
			GetInfoFunc: func(ctx context.Context) (*truenas.SystemInfo, error) {
				return &truenas.SystemInfo{
					Hostname:      "truenas",
					Model:         "AMD Ryzen 5 2400G",
					Cores:         8,
					PhysicalCores: 4,
					UptimeSeconds: 259200,
				}, nil
			},
			GetVersionFunc: func(ctx context.Context) (string, error) {
				return "TrueNAS-25.04.0", nil
			},
		},
		&truenas.MockReportingService{},
		&truenas.MockInterfaceService{
			ListFunc: func(ctx context.Context) ([]truenas.NetworkInterface, error) {
				return []truenas.NetworkInterface{
					{
						ID: "enp24s0", Name: "enp24s0",
						Type:  truenas.InterfaceTypePhysical,
						State: truenas.InterfaceState{LinkState: truenas.LinkStateUp},
					},
				}, nil
			},
		},
		&truenas.MockAppService{
			ListAppsFunc: func(ctx context.Context) ([]truenas.App, error) {
				return []truenas.App{
					{Name: "tailscale", State: "RUNNING"},
					{Name: "sonarr", State: "RUNNING"},
				}, nil
			},
		},
	)
}

func newApp(svc *internal.Services) *app.App {
	return app.New(app.Params{Services: svc, ServerName: "test-server", StaleTTL: testStaleTTL})
}

func TestApp_New(t *testing.T) {
	a := newApp(newTestServices())
	if a == nil {
		t.Fatal("expected non-nil app")
	}
}

func TestApp_New_WithServices(t *testing.T) {
	a := newApp(newTestServicesWithData())
	if !a.IsConnected() {
		t.Error("expected connected when Services provided")
	}
}

func TestApp_New_WithoutServices(t *testing.T) {
	a := app.New(app.Params{ServerName: "test-server", StaleTTL: testStaleTTL})
	if a.IsConnected() {
		t.Error("expected not connected when no Services or Connect provided")
	}
}

func TestApp_ActiveTab(t *testing.T) {
	a := newApp(newTestServices())
	if a.ActiveTab() != 0 {
		t.Errorf("expected initial tab 0, got %d", a.ActiveTab())
	}
}

func TestApp_SetTab(t *testing.T) {
	a := newApp(newTestServices())
	a.SetTab(1)
	if a.ActiveTab() != 1 {
		t.Errorf("expected tab 1, got %d", a.ActiveTab())
	}
	a.SetTab(3)
	if a.ActiveTab() != 3 {
		t.Errorf("expected tab 3, got %d", a.ActiveTab())
	}
}

func TestApp_ServerName(t *testing.T) {
	a := app.New(app.Params{Services: newTestServices(), ServerName: "home", StaleTTL: testStaleTTL})
	if a.ServerName() != "home" {
		t.Errorf("expected server name home, got %s", a.ServerName())
	}
}

func TestApp_LoadActiveView_Tab0_Dashboard(t *testing.T) {
	svc := newTestServicesWithData()
	a := newApp(svc)
	a.SetTab(0)

	err := a.LoadActiveView(context.Background())
	if err != nil {
		t.Fatalf("unexpected error loading dashboard: %v", err)
	}
}

func TestApp_LoadActiveView_Tab1_Pools(t *testing.T) {
	svc := newTestServicesWithData()
	a := newApp(svc)
	a.SetTab(1)

	err := a.LoadActiveView(context.Background())
	if err != nil {
		t.Fatalf("unexpected error loading pools: %v", err)
	}
}

func TestApp_LoadActiveView_Tab2_Datasets(t *testing.T) {
	svc := newTestServicesWithData()
	a := newApp(svc)
	a.SetTab(2)

	err := a.LoadActiveView(context.Background())
	if err != nil {
		t.Fatalf("unexpected error loading datasets: %v", err)
	}
}

func TestApp_LoadActiveView_Tab3_Snapshots(t *testing.T) {
	svc := newTestServicesWithData()
	a := newApp(svc)
	a.SetTab(3)

	err := a.LoadActiveView(context.Background())
	if err != nil {
		t.Fatalf("unexpected error loading snapshots: %v", err)
	}
}

func TestApp_LoadActiveView_NotConnected(t *testing.T) {
	a := app.New(app.Params{ServerName: "test-server", StaleTTL: testStaleTTL})
	err := a.LoadActiveView(context.Background())
	if err != nil {
		t.Fatalf("expected nil error when not connected, got %v", err)
	}
}

func TestApp_LoadActiveView_Error_Pools(t *testing.T) {
	svc := internal.NewServices(
		&truenas.MockDatasetService{
			ListPoolsFunc: func(ctx context.Context) ([]truenas.Pool, error) {
				return nil, context.DeadlineExceeded
			},
		},
		&truenas.MockSnapshotService{},
		&truenas.MockSystemService{
			GetInfoFunc: func(ctx context.Context) (*truenas.SystemInfo, error) {
				return &truenas.SystemInfo{Hostname: "test", Model: "Test"}, nil
			},
			GetVersionFunc: func(ctx context.Context) (string, error) { return "v1", nil },
		},
		&truenas.MockReportingService{},
		&truenas.MockInterfaceService{
			ListFunc: func(ctx context.Context) ([]truenas.NetworkInterface, error) { return nil, nil },
		},
		&truenas.MockAppService{
			ListAppsFunc: func(ctx context.Context) ([]truenas.App, error) { return nil, nil },
		},
	)
	a := newApp(svc)
	a.SetTab(1)

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
		&truenas.MockSystemService{
			GetInfoFunc: func(ctx context.Context) (*truenas.SystemInfo, error) {
				return &truenas.SystemInfo{Hostname: "test", Model: "Test"}, nil
			},
			GetVersionFunc: func(ctx context.Context) (string, error) { return "v1", nil },
		},
		&truenas.MockReportingService{},
		&truenas.MockInterfaceService{
			ListFunc: func(ctx context.Context) ([]truenas.NetworkInterface, error) { return nil, nil },
		},
		&truenas.MockAppService{
			ListAppsFunc: func(ctx context.Context) ([]truenas.App, error) { return nil, nil },
		},
	)
	a := newApp(svc)
	a.SetTab(2)

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
		&truenas.MockSystemService{
			GetInfoFunc: func(ctx context.Context) (*truenas.SystemInfo, error) {
				return &truenas.SystemInfo{Hostname: "test", Model: "Test"}, nil
			},
			GetVersionFunc: func(ctx context.Context) (string, error) { return "v1", nil },
		},
		&truenas.MockReportingService{},
		&truenas.MockInterfaceService{
			ListFunc: func(ctx context.Context) ([]truenas.NetworkInterface, error) { return nil, nil },
		},
		&truenas.MockAppService{
			ListAppsFunc: func(ctx context.Context) ([]truenas.App, error) { return nil, nil },
		},
	)
	a := newApp(svc)
	a.SetTab(3)

	err := a.LoadActiveView(context.Background())
	if err == nil {
		t.Fatal("expected error to propagate from snapshots Load")
	}
}

func TestApp_Draw(t *testing.T) {
	svc := newTestServicesWithData()
	a := newApp(svc)
	_ = a.LoadActiveView(context.Background())

	ctx := testDrawContext(100, 30)
	s, err := a.Draw(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if s.Size.Width != 100 {
		t.Errorf("expected surface width=100, got %d", s.Size.Width)
	}
	if s.Size.Height != 30 {
		t.Errorf("expected surface height=30, got %d", s.Size.Height)
	}
}

func TestApp_Draw_AllTabs(t *testing.T) {
	svc := newTestServicesWithData()
	a := newApp(svc)

	for tab := 0; tab < 4; tab++ {
		a.SetTab(tab)
		_ = a.LoadActiveView(context.Background())

		ctx := testDrawContext(100, 30)
		_, err := a.Draw(ctx)
		if err != nil {
			t.Fatalf("unexpected error drawing tab %d: %v", tab, err)
		}
	}
}

func TestApp_Draw_BeforeLoad(t *testing.T) {
	a := newApp(newTestServices())

	ctx := testDrawContext(100, 30)
	_, err := a.Draw(ctx)
	if err != nil {
		t.Fatalf("unexpected error drawing before load: %v", err)
	}
}

func TestApp_Draw_Connecting(t *testing.T) {
	a := app.New(app.Params{ServerName: "nas-1", StaleTTL: testStaleTTL})

	ctx := testDrawContext(80, 24)
	s, err := a.Draw(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if s.Size.Width != 80 {
		t.Errorf("expected surface width=80, got %d", s.Size.Width)
	}
}

func TestApp_Draw_ConnectFailed(t *testing.T) {
	a := app.New(app.Params{ServerName: "nas-1", StaleTTL: testStaleTTL})

	// Simulate connect failure
	_, _ = a.HandleEvent(app.ConnectFailed{Err: fmt.Errorf("connection refused")}, vxfw.EventPhase(0))

	ctx := testDrawContext(80, 24)
	s, err := a.Draw(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if s.Size.Width != 80 {
		t.Errorf("expected surface width=80, got %d", s.Size.Width)
	}
}

func TestApp_CaptureEvent_Quit(t *testing.T) {
	a := newApp(newTestServices())

	cmd, err := a.CaptureEvent(vaxis.Key{Keycode: 'q'})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, ok := cmd.(vxfw.QuitCmd); !ok {
		t.Errorf("expected QuitCmd, got %T", cmd)
	}
}

func TestApp_CaptureEvent_QuitWhenNotConnected(t *testing.T) {
	a := app.New(app.Params{ServerName: "test-server", StaleTTL: testStaleTTL})

	cmd, err := a.CaptureEvent(vaxis.Key{Keycode: 'q'})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, ok := cmd.(vxfw.QuitCmd); !ok {
		t.Errorf("expected QuitCmd even when not connected, got %T", cmd)
	}
}

func TestApp_CaptureEvent_IgnoredWhenNotConnected(t *testing.T) {
	a := app.New(app.Params{ServerName: "test-server", StaleTTL: testStaleTTL})

	for _, key := range []rune{'r', '1', '2', '3', '4'} {
		cmd, err := a.CaptureEvent(vaxis.Key{Keycode: key})
		if err != nil {
			t.Fatalf("unexpected error for key '%c': %v", key, err)
		}
		if cmd != nil {
			t.Errorf("expected nil command for key '%c' when not connected, got %T", key, cmd)
		}
	}
}

func TestApp_CaptureEvent_NumberKeys(t *testing.T) {
	a := newApp(newTestServices())

	tests := []struct {
		key      rune
		expected int
	}{
		{'1', 0},
		{'2', 1},
		{'3', 2},
		{'4', 3},
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
	a := newApp(newTestServices())

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
	a := newApp(newTestServices())

	cmd, err := a.CaptureEvent(vaxis.Key{Keycode: vaxis.KeyTab, Modifiers: vaxis.ModShift})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cmd == nil {
		t.Fatal("expected non-nil command for Shift+Tab")
	}
	if a.ActiveTab() != 3 {
		t.Errorf("expected tab 3 after Shift+Tab, got %d", a.ActiveTab())
	}
}

func TestApp_CaptureEvent_UnhandledKey(t *testing.T) {
	a := newApp(newTestServices())

	cmd, err := a.CaptureEvent(vaxis.Key{Keycode: 'x'})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cmd != nil {
		t.Errorf("expected nil command for unhandled key, got %T", cmd)
	}
}

func TestApp_CaptureEvent_NonKeyEvent(t *testing.T) {
	a := newApp(newTestServices())

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
	a := newApp(svc)

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
	a := newApp(svc)
	_ = a.LoadActiveView(context.Background())

	cmd, err := a.HandleEvent(vaxis.Key{Keycode: 'j'}, vxfw.EventPhase(0))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	_ = cmd
}

func TestApp_HandleEvent_AllTabs(t *testing.T) {
	svc := newTestServicesWithData()
	a := newApp(svc)

	for tab := 0; tab < 4; tab++ {
		a.SetTab(tab)
		_ = a.LoadActiveView(context.Background())

		cmd, err := a.HandleEvent(vaxis.Key{Keycode: 'j'}, vxfw.EventPhase(0))
		if err != nil {
			t.Fatalf("unexpected error on tab %d: %v", tab, err)
		}
		_ = cmd
	}
}

func TestApp_HandleEvent_Init_WithConnectFn(t *testing.T) {
	svc := newTestServicesWithData()
	called := false

	a := app.New(app.Params{
		ServerName: "test-server",
		StaleTTL:   testStaleTTL,
		Connect: func(ctx context.Context) (*internal.Services, error) {
			called = true
			return svc, nil
		},
	})

	done := make(chan struct{}, 1)
	a.SetPostEvent(func(ev vaxis.Event) {
		if _, ok := ev.(app.Connected); ok {
			done <- struct{}{}
		}
	})

	_, err := a.HandleEvent(vxfw.Init{}, vxfw.EventPhase(0))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("timed out waiting for connect callback")
	}

	if !called {
		t.Error("expected Connect callback to be called")
	}
}

func TestApp_HandleEvent_Init_WithConnectFn_Error(t *testing.T) {
	a := app.New(app.Params{
		ServerName: "test-server",
		StaleTTL:   testStaleTTL,
		Connect: func(ctx context.Context) (*internal.Services, error) {
			return nil, fmt.Errorf("connection refused")
		},
	})

	done := make(chan struct{}, 1)
	a.SetPostEvent(func(ev vaxis.Event) {
		if _, ok := ev.(app.ConnectFailed); ok {
			done <- struct{}{}
		}
	})

	_, _ = a.HandleEvent(vxfw.Init{}, vxfw.EventPhase(0))

	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("timed out waiting for ConnectFailed event")
	}
}

func TestApp_HandleEvent_Init_NoConnectFn(t *testing.T) {
	a := newApp(newTestServices())

	cmd, err := a.HandleEvent(vxfw.Init{}, vxfw.EventPhase(0))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cmd != nil {
		t.Errorf("expected nil command from Init without connectFn, got %T", cmd)
	}
}

func TestApp_HandleEvent_Connected(t *testing.T) {
	a := app.New(app.Params{ServerName: "test-server", StaleTTL: testStaleTTL})

	var mu sync.Mutex
	var events []views.ViewLoaded
	done := make(chan struct{}, 4)

	a.SetPostEvent(func(ev vaxis.Event) {
		if vl, ok := ev.(views.ViewLoaded); ok {
			mu.Lock()
			events = append(events, vl)
			mu.Unlock()
			done <- struct{}{}
		}
	})

	svc := newTestServicesWithData()
	cmd, err := a.HandleEvent(app.Connected{Services: svc}, vxfw.EventPhase(0))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, ok := cmd.(vxfw.RedrawCmd); !ok {
		t.Errorf("expected RedrawCmd, got %T", cmd)
	}
	if !a.IsConnected() {
		t.Error("expected connected after Connected event")
	}

	// Wait for LoadAll goroutines (4 tabs now)
	for i := 0; i < 4; i++ {
		select {
		case <-done:
		case <-time.After(5 * time.Second):
			t.Fatal("timed out waiting for LoadAll after Connected")
		}
	}

	mu.Lock()
	defer mu.Unlock()
	if len(events) != 4 {
		t.Fatalf("expected 4 ViewLoaded events, got %d", len(events))
	}
}

func TestApp_HandleEvent_ConnectFailed(t *testing.T) {
	a := app.New(app.Params{ServerName: "test-server", StaleTTL: testStaleTTL})

	cmd, err := a.HandleEvent(app.ConnectFailed{Err: fmt.Errorf("refused")}, vxfw.EventPhase(0))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, ok := cmd.(vxfw.RedrawCmd); !ok {
		t.Errorf("expected RedrawCmd, got %T", cmd)
	}
	if a.IsConnected() {
		t.Error("expected not connected after ConnectFailed")
	}
}

func TestApp_HandleEvent_ViewLoaded(t *testing.T) {
	a := newApp(newTestServices())

	cmd, err := a.HandleEvent(views.ViewLoaded{Tab: 1, Err: nil}, vxfw.EventPhase(0))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, ok := cmd.(vxfw.RedrawCmd); !ok {
		t.Errorf("expected RedrawCmd, got %T", cmd)
	}
}

func TestApp_HandleEvent_ViewLoaded_WithError(t *testing.T) {
	a := newApp(newTestServices())

	cmd, err := a.HandleEvent(views.ViewLoaded{Tab: 2, Err: context.DeadlineExceeded}, vxfw.EventPhase(0))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, ok := cmd.(vxfw.RedrawCmd); !ok {
		t.Errorf("expected RedrawCmd even on load error, got %T", cmd)
	}
}

func TestApp_HandleEvent_DashboardUpdated(t *testing.T) {
	a := newApp(newTestServices())

	cmd, err := a.HandleEvent(views.DashboardUpdated{}, vxfw.EventPhase(0))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, ok := cmd.(vxfw.RedrawCmd); !ok {
		t.Errorf("expected RedrawCmd for DashboardUpdated, got %T", cmd)
	}
}

func TestApp_HandleEvent_NotConnected(t *testing.T) {
	a := app.New(app.Params{ServerName: "test-server", StaleTTL: testStaleTTL})

	// Key events to activeView should not panic when not connected
	cmd, err := a.HandleEvent(vaxis.Key{Keycode: 'j'}, vxfw.EventPhase(0))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cmd != nil {
		t.Errorf("expected nil command when not connected, got %T", cmd)
	}
}

func TestApp_LoadAll(t *testing.T) {
	svc := newTestServicesWithData()
	a := newApp(svc)

	var mu sync.Mutex
	var events []views.ViewLoaded
	done := make(chan struct{}, 4)

	a.SetPostEvent(func(ev vaxis.Event) {
		if vl, ok := ev.(views.ViewLoaded); ok {
			mu.Lock()
			events = append(events, vl)
			mu.Unlock()
			done <- struct{}{}
		}
	})

	a.LoadAll(context.Background())

	for i := 0; i < 4; i++ {
		select {
		case <-done:
		case <-time.After(5 * time.Second):
			t.Fatal("timed out waiting for LoadAll to complete")
		}
	}

	mu.Lock()
	defer mu.Unlock()

	if len(events) != 4 {
		t.Fatalf("expected 4 ViewLoaded events, got %d", len(events))
	}

	tabs := map[int]bool{}
	for _, ev := range events {
		if ev.Err != nil {
			t.Errorf("tab %d had unexpected error: %v", ev.Tab, ev.Err)
		}
		tabs[ev.Tab] = true
	}
	for i := 0; i < 4; i++ {
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
		&truenas.MockSystemService{
			GetInfoFunc: func(ctx context.Context) (*truenas.SystemInfo, error) {
				return &truenas.SystemInfo{Hostname: "test", Model: "Test"}, nil
			},
			GetVersionFunc: func(ctx context.Context) (string, error) { return "v1", nil },
		},
		&truenas.MockReportingService{},
		&truenas.MockInterfaceService{
			ListFunc: func(ctx context.Context) ([]truenas.NetworkInterface, error) { return nil, nil },
		},
		&truenas.MockAppService{
			ListAppsFunc: func(ctx context.Context) ([]truenas.App, error) { return nil, nil },
		},
	)
	a := newApp(svc)

	var mu sync.Mutex
	var events []views.ViewLoaded
	done := make(chan struct{}, 4)

	a.SetPostEvent(func(ev vaxis.Event) {
		if vl, ok := ev.(views.ViewLoaded); ok {
			mu.Lock()
			events = append(events, vl)
			mu.Unlock()
			done <- struct{}{}
		}
	})

	a.LoadAll(context.Background())

	for i := 0; i < 4; i++ {
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

func TestApp_LoadAll_NotConnected(t *testing.T) {
	a := app.New(app.Params{ServerName: "test-server", StaleTTL: testStaleTTL})
	// Should not panic
	a.LoadAll(context.Background())
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
		&truenas.MockSystemService{
			GetInfoFunc: func(ctx context.Context) (*truenas.SystemInfo, error) {
				return &truenas.SystemInfo{Hostname: "test", Model: "Test"}, nil
			},
			GetVersionFunc: func(ctx context.Context) (string, error) { return "v1", nil },
		},
		&truenas.MockReportingService{},
		&truenas.MockInterfaceService{
			ListFunc: func(ctx context.Context) ([]truenas.NetworkInterface, error) { return nil, nil },
		},
		&truenas.MockAppService{
			ListAppsFunc: func(ctx context.Context) ([]truenas.App, error) { return nil, nil },
		},
	)

	done := make(chan struct{}, 1)
	a := app.New(app.Params{Services: svc, ServerName: "test-server", StaleTTL: 0})
	a.SetPostEvent(func(ev vaxis.Event) {
		done <- struct{}{}
	})
	a.SetTab(2) // datasets tab
	_ = a.LoadActiveView(context.Background())
	initial := callCount

	time.Sleep(time.Millisecond)

	a.SetTab(0) // switch away
	_, _ = a.CaptureEvent(vaxis.Key{Keycode: '3'}) // switch to datasets via key

	<-done

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
		&truenas.MockSystemService{
			GetInfoFunc: func(ctx context.Context) (*truenas.SystemInfo, error) {
				return &truenas.SystemInfo{Hostname: "test", Model: "Test"}, nil
			},
			GetVersionFunc: func(ctx context.Context) (string, error) { return "v1", nil },
		},
		&truenas.MockReportingService{},
		&truenas.MockInterfaceService{
			ListFunc: func(ctx context.Context) ([]truenas.NetworkInterface, error) { return nil, nil },
		},
		&truenas.MockAppService{
			ListAppsFunc: func(ctx context.Context) ([]truenas.App, error) { return nil, nil },
		},
	)

	a := app.New(app.Params{Services: svc, ServerName: "test-server", StaleTTL: time.Hour})
	a.SetTab(2) // datasets
	_ = a.LoadActiveView(context.Background())
	afterLoad := callCount

	a.SetTab(0) // switch away
	_, _ = a.CaptureEvent(vaxis.Key{Keycode: '3'}) // switch back to datasets

	if callCount != afterLoad {
		t.Errorf("expected no refetch when fresh, but ListDatasetsFunc called %d more times", callCount-afterLoad)
	}
}

func TestApp_DashboardNeverStale(t *testing.T) {
	svc := newTestServicesWithData()
	a := app.New(app.Params{Services: svc, ServerName: "test-server", StaleTTL: 0})

	refetchCount := 0
	a.SetPostEvent(func(ev vaxis.Event) {
		if _, ok := ev.(views.ViewLoaded); ok {
			refetchCount++
		}
	})

	// Load dashboard
	_ = a.LoadActiveView(context.Background())
	time.Sleep(time.Millisecond)

	// Switch away and back to dashboard
	a.SetTab(1)
	_, _ = a.CaptureEvent(vaxis.Key{Keycode: '1'}) // back to dashboard

	// Dashboard should not trigger a refetch even with StaleTTL=0
	time.Sleep(10 * time.Millisecond)
	if refetchCount > 0 {
		t.Error("dashboard should never be refetched via stale mechanism")
	}
}
