package views_test

import (
	"context"
	"testing"
	"time"

	"git.sr.ht/~rockorager/vaxis"
	"git.sr.ht/~rockorager/vaxis/vxfw"
	"github.com/deevus/truenas-go"
	"github.com/deevus/truenas-tui/views"
)

func mockDashboardServices() views.DashboardViewParams {
	return views.DashboardViewParams{
		System: &truenas.MockSystemService{
			GetInfoFunc: func(ctx context.Context) (*truenas.SystemInfo, error) {
				return &truenas.SystemInfo{
					Hostname:      "truenas",
					Model:         "AMD Ryzen 5 2400G",
					Cores:         8,
					PhysicalCores: 4,
					Uptime:        "3 days",
					UptimeSeconds: 259200,
					LoadAvg:       [3]float64{1.0, 0.8, 0.5},
				}, nil
			},
			GetVersionFunc: func(ctx context.Context) (string, error) {
				return "TrueNAS-25.04.0", nil
			},
		},
		Reporting: &truenas.MockReportingService{},
		Interfaces: &truenas.MockInterfaceService{
			ListFunc: func(ctx context.Context) ([]truenas.NetworkInterface, error) {
				return []truenas.NetworkInterface{
					{
						ID:   "enp24s0",
						Name: "enp24s0",
						Type: truenas.InterfaceTypePhysical,
						State: truenas.InterfaceState{
							LinkState: truenas.LinkStateUp,
						},
					},
				}, nil
			},
		},
		Apps: &truenas.MockAppService{
			ListAppsFunc: func(ctx context.Context) ([]truenas.App, error) {
				return []truenas.App{
					{Name: "tailscale", State: "RUNNING"},
					{Name: "sonarr", State: "RUNNING"},
					{Name: "stopped-app", State: "STOPPED"},
				}, nil
			},
		},
	}
}

func TestDashboardView_New(t *testing.T) {
	dv := views.NewDashboardView(mockDashboardServices())
	if dv == nil {
		t.Fatal("expected non-nil DashboardView")
	}
	if dv.Loaded() {
		t.Error("expected not loaded before Load()")
	}
}

func TestDashboardView_Load(t *testing.T) {
	dv := views.NewDashboardView(mockDashboardServices())
	err := dv.Load(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !dv.Loaded() {
		t.Error("expected loaded after Load()")
	}
}

func TestDashboardView_Load_Error(t *testing.T) {
	params := mockDashboardServices()
	params.System = &truenas.MockSystemService{
		GetInfoFunc: func(ctx context.Context) (*truenas.SystemInfo, error) {
			return nil, context.DeadlineExceeded
		},
	}
	dv := views.NewDashboardView(params)
	err := dv.Load(context.Background())
	if err == nil {
		t.Fatal("expected error from Load()")
	}
	if dv.Loaded() {
		t.Error("should not be loaded on error")
	}
}

func TestDashboardView_Draw_BeforeLoad(t *testing.T) {
	dv := views.NewDashboardView(mockDashboardServices())
	ctx := testDrawContext(80, 24)
	_, err := dv.Draw(ctx)
	if err != nil {
		t.Fatalf("unexpected error drawing before load: %v", err)
	}
}

func TestDashboardView_Draw_AfterLoad(t *testing.T) {
	dv := views.NewDashboardView(mockDashboardServices())
	if err := dv.Load(context.Background()); err != nil {
		t.Fatalf("Load: %v", err)
	}

	ctx := testDrawContext(100, 30)
	s, err := dv.Draw(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if s.Size.Width != 100 {
		t.Errorf("expected width=100, got %d", s.Size.Width)
	}
	if s.Size.Height != 30 {
		t.Errorf("expected height=30, got %d", s.Size.Height)
	}
}

func TestDashboardView_HandleEvent(t *testing.T) {
	dv := views.NewDashboardView(mockDashboardServices())
	_ = dv.Load(context.Background())

	_, err := dv.HandleEvent(vaxis.Key{Keycode: 'j'}, vxfw.EventPhase(0))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestDashboardView_StartStopSubscriptions(t *testing.T) {
	realtimeCh := make(chan truenas.RealtimeUpdate, 1)
	statsCh := make(chan []truenas.AppStats, 1)

	params := mockDashboardServices()
	var events []views.DashboardUpdated
	params.PostEvent = func(ev vaxis.Event) {
		if du, ok := ev.(views.DashboardUpdated); ok {
			events = append(events, du)
		}
	}
	params.Reporting = &truenas.MockReportingService{
		SubscribeRealtimeFunc: func(ctx context.Context) (*truenas.Subscription[truenas.RealtimeUpdate], error) {
			return truenas.NewSubscription(realtimeCh, func() { close(realtimeCh) }), nil
		},
	}
	params.Apps = &truenas.MockAppService{
		ListAppsFunc: func(ctx context.Context) ([]truenas.App, error) {
			return []truenas.App{{Name: "test", State: "RUNNING"}}, nil
		},
		SubscribeStatsFunc: func(ctx context.Context) (*truenas.Subscription[[]truenas.AppStats], error) {
			return truenas.NewSubscription(statsCh, func() { close(statsCh) }), nil
		},
	}

	dv := views.NewDashboardView(params)
	if err := dv.Load(context.Background()); err != nil {
		t.Fatalf("Load: %v", err)
	}

	dv.StartSubscriptions(context.Background())

	// Send a realtime update
	realtimeCh <- truenas.RealtimeUpdate{
		CPU: map[string]truenas.RealtimeCPU{
			"0": {Usage: 50, Temperature: 65},
		},
		Memory: truenas.RealtimeMemory{
			PhysicalTotal:     16 * 1024 * 1024 * 1024,
			PhysicalAvailable: 4 * 1024 * 1024 * 1024,
		},
	}

	// Send app stats
	statsCh <- []truenas.AppStats{
		{AppName: "test", CPUUsage: 1.5, Memory: 100 * 1024 * 1024},
	}

	// Wait briefly for goroutines to process
	time.Sleep(50 * time.Millisecond)

	dv.StopSubscriptions()

	// After receiving data, Draw should still work
	ctx := testDrawContext(100, 30)
	_, err := dv.Draw(ctx)
	if err != nil {
		t.Fatalf("unexpected error drawing after subscriptions: %v", err)
	}
}

// blockingSub returns a subscription that blocks until ctx is cancelled.
func blockingRealtimeSub(ctx context.Context) (*truenas.Subscription[truenas.RealtimeUpdate], error) {
	ch := make(chan truenas.RealtimeUpdate)
	go func() { <-ctx.Done(); close(ch) }()
	return truenas.NewSubscription((<-chan truenas.RealtimeUpdate)(ch), func() {}), nil
}

func blockingStatsSub(ctx context.Context) (*truenas.Subscription[[]truenas.AppStats], error) {
	ch := make(chan []truenas.AppStats)
	go func() { <-ctx.Done(); close(ch) }()
	return truenas.NewSubscription((<-chan []truenas.AppStats)(ch), func() {}), nil
}

func TestDashboardView_RealtimeSub_RetryOnError(t *testing.T) {
	var attempt int
	realtimeCh := make(chan truenas.RealtimeUpdate, 1)

	params := mockDashboardServices()
	updated := make(chan struct{}, 5)
	params.PostEvent = func(ev vaxis.Event) {
		if _, ok := ev.(views.DashboardUpdated); ok {
			select {
			case updated <- struct{}{}:
			default:
			}
		}
	}
	params.Reporting = &truenas.MockReportingService{
		SubscribeRealtimeFunc: func(ctx context.Context) (*truenas.Subscription[truenas.RealtimeUpdate], error) {
			attempt++
			if attempt <= 2 {
				return nil, context.DeadlineExceeded
			}
			return truenas.NewSubscription(realtimeCh, func() {}), nil
		},
	}
	params.Apps = &truenas.MockAppService{
		ListAppsFunc:       params.Apps.(*truenas.MockAppService).ListAppsFunc,
		SubscribeStatsFunc: blockingStatsSub,
	}

	dv := views.NewDashboardView(params)
	dv.RetryBaseDelay = time.Millisecond
	if err := dv.Load(context.Background()); err != nil {
		t.Fatalf("Load: %v", err)
	}

	dv.StartSubscriptions(context.Background())
	defer dv.StopSubscriptions()

	// After retries, the subscription should work
	realtimeCh <- truenas.RealtimeUpdate{
		CPU: map[string]truenas.RealtimeCPU{
			"0": {Usage: 42},
		},
	}

	select {
	case <-updated:
		// success — received update after retry
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for realtime update after retry")
	}

	if attempt < 3 {
		t.Errorf("expected at least 3 attempts, got %d", attempt)
	}
}

func TestDashboardView_StatsSub_RetryOnError(t *testing.T) {
	var attempt int
	statsCh := make(chan []truenas.AppStats, 1)

	params := mockDashboardServices()
	updated := make(chan struct{}, 5)
	params.PostEvent = func(ev vaxis.Event) {
		if _, ok := ev.(views.DashboardUpdated); ok {
			select {
			case updated <- struct{}{}:
			default:
			}
		}
	}
	params.Reporting = &truenas.MockReportingService{
		SubscribeRealtimeFunc: blockingRealtimeSub,
	}
	params.Apps = &truenas.MockAppService{
		ListAppsFunc: func(ctx context.Context) ([]truenas.App, error) {
			return []truenas.App{{Name: "test", State: "RUNNING"}}, nil
		},
		SubscribeStatsFunc: func(ctx context.Context) (*truenas.Subscription[[]truenas.AppStats], error) {
			attempt++
			if attempt <= 2 {
				return nil, context.DeadlineExceeded
			}
			return truenas.NewSubscription(statsCh, func() {}), nil
		},
	}

	dv := views.NewDashboardView(params)
	dv.RetryBaseDelay = time.Millisecond
	if err := dv.Load(context.Background()); err != nil {
		t.Fatalf("Load: %v", err)
	}

	dv.StartSubscriptions(context.Background())
	defer dv.StopSubscriptions()

	statsCh <- []truenas.AppStats{
		{AppName: "test", CPUUsage: 5.0, Memory: 200 * 1024 * 1024},
	}

	select {
	case <-updated:
		// success
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for stats update after retry")
	}

	if attempt < 3 {
		t.Errorf("expected at least 3 attempts, got %d", attempt)
	}
}

func TestDashboardView_RealtimeSub_ReconnectOnClose(t *testing.T) {
	callCount := 0
	ch1 := make(chan truenas.RealtimeUpdate, 1)
	ch2 := make(chan truenas.RealtimeUpdate, 1)

	params := mockDashboardServices()
	updated := make(chan struct{}, 5)
	params.PostEvent = func(ev vaxis.Event) {
		if _, ok := ev.(views.DashboardUpdated); ok {
			select {
			case updated <- struct{}{}:
			default:
			}
		}
	}
	params.Reporting = &truenas.MockReportingService{
		SubscribeRealtimeFunc: func(ctx context.Context) (*truenas.Subscription[truenas.RealtimeUpdate], error) {
			callCount++
			if callCount == 1 {
				return truenas.NewSubscription((<-chan truenas.RealtimeUpdate)(ch1), func() {}), nil
			}
			return truenas.NewSubscription((<-chan truenas.RealtimeUpdate)(ch2), func() {}), nil
		},
	}
	params.Apps = &truenas.MockAppService{
		ListAppsFunc:       params.Apps.(*truenas.MockAppService).ListAppsFunc,
		SubscribeStatsFunc: blockingStatsSub,
	}

	dv := views.NewDashboardView(params)
	dv.RetryBaseDelay = time.Millisecond
	if err := dv.Load(context.Background()); err != nil {
		t.Fatalf("Load: %v", err)
	}

	dv.StartSubscriptions(context.Background())
	defer dv.StopSubscriptions()

	// Send on first channel, verify receipt
	ch1 <- truenas.RealtimeUpdate{CPU: map[string]truenas.RealtimeCPU{"0": {Usage: 10}}}
	select {
	case <-updated:
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for first update")
	}

	// Close first channel to trigger reconnect
	close(ch1)

	// Give reconnect time (backoff is 1ms)
	time.Sleep(20 * time.Millisecond)

	// Send on second channel
	ch2 <- truenas.RealtimeUpdate{CPU: map[string]truenas.RealtimeCPU{"0": {Usage: 20}}}
	select {
	case <-updated:
		// success — reconnected and received data
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for update after reconnect")
	}

	if callCount < 2 {
		t.Errorf("expected at least 2 subscribe calls, got %d", callCount)
	}
}

func TestDashboardView_Sub_CancelDuringBackoff(t *testing.T) {
	params := mockDashboardServices()
	params.Reporting = &truenas.MockReportingService{
		SubscribeRealtimeFunc: func(ctx context.Context) (*truenas.Subscription[truenas.RealtimeUpdate], error) {
			return nil, context.DeadlineExceeded
		},
	}
	params.Apps = &truenas.MockAppService{
		ListAppsFunc: func(ctx context.Context) ([]truenas.App, error) {
			return []truenas.App{{Name: "test", State: "RUNNING"}}, nil
		},
		SubscribeStatsFunc: func(ctx context.Context) (*truenas.Subscription[[]truenas.AppStats], error) {
			return nil, context.DeadlineExceeded
		},
	}

	dv := views.NewDashboardView(params)
	dv.RetryBaseDelay = time.Second // long enough that we cancel during it
	if err := dv.Load(context.Background()); err != nil {
		t.Fatalf("Load: %v", err)
	}

	done := make(chan struct{})
	ctx, cancel := context.WithCancel(context.Background())
	dv.StartSubscriptions(ctx)

	go func() {
		// Wait for first retry attempt to start backing off
		time.Sleep(50 * time.Millisecond)
		cancel()
		dv.StopSubscriptions()
		close(done)
	}()

	select {
	case <-done:
		// Goroutines exited cleanly
	case <-time.After(3 * time.Second):
		t.Fatal("goroutines did not exit after context cancellation")
	}
}

func TestFormatUptime(t *testing.T) {
	tests := []struct {
		seconds float64
		want    string
	}{
		{259200, "3d 0h"},          // 3 days
		{90061, "1d 1h"},           // 1 day 1 hour
		{7200, "2h 0m"},            // 2 hours
		{3661, "1h 1m"},            // 1 hour 1 minute
		{300, "5m"},                // 5 minutes
		{59, "0m"},                 // less than 1 minute
	}
	for _, tt := range tests {
		got := views.FormatUptime(tt.seconds)
		if got != tt.want {
			t.Errorf("FormatUptime(%v) = %q, want %q", tt.seconds, got, tt.want)
		}
	}
}

func TestDashboardView_Draw_ShortUptime(t *testing.T) {
	params := mockDashboardServices()
	params.System = &truenas.MockSystemService{
		GetInfoFunc: func(ctx context.Context) (*truenas.SystemInfo, error) {
			return &truenas.SystemInfo{
				Hostname:      "truenas",
				Model:         "Test CPU",
				UptimeSeconds: 300, // 5 minutes, exercises minutes-only branch
			}, nil
		},
		GetVersionFunc: func(ctx context.Context) (string, error) {
			return "TrueNAS-25.04.0", nil
		},
	}

	dv := views.NewDashboardView(params)
	if err := dv.Load(context.Background()); err != nil {
		t.Fatalf("Load: %v", err)
	}

	ctx := testDrawContext(100, 30)
	_, err := dv.Draw(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
