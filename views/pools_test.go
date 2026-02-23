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

func newPoolsView(mock *truenas.MockDatasetService) *views.PoolsView {
	return views.NewPoolsView(views.PoolsViewParams{
		Service:  mock,
		StaleTTL: 30 * time.Second,
	})
}

func TestPoolsView_Load(t *testing.T) {
	mock := &truenas.MockDatasetService{
		ListPoolsFunc: func(ctx context.Context) ([]truenas.Pool, error) {
			return []truenas.Pool{
				{ID: 1, Name: "tank", Status: "ONLINE", Size: 1099511627776, Allocated: 549755813888, Free: 549755813888},
				{ID: 2, Name: "backup", Status: "ONLINE", Size: 2199023255552, Allocated: 0, Free: 2199023255552},
			}, nil
		},
	}

	pv := newPoolsView(mock)
	err := pv.Load(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	pools := pv.Pools()
	if len(pools) != 2 {
		t.Fatalf("expected 2 pools, got %d", len(pools))
	}
	if pools[0].Name != "tank" {
		t.Errorf("expected first pool name=tank, got %s", pools[0].Name)
	}
}

func TestPoolsView_Load_Error(t *testing.T) {
	mock := &truenas.MockDatasetService{
		ListPoolsFunc: func(ctx context.Context) ([]truenas.Pool, error) {
			return nil, context.DeadlineExceeded
		},
	}

	pv := newPoolsView(mock)
	err := pv.Load(context.Background())
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestPoolsView_ItemCount(t *testing.T) {
	mock := &truenas.MockDatasetService{
		ListPoolsFunc: func(ctx context.Context) ([]truenas.Pool, error) {
			return []truenas.Pool{
				{ID: 1, Name: "tank", Status: "ONLINE"},
			}, nil
		},
	}

	pv := newPoolsView(mock)
	if pv.ItemCount() != 0 {
		t.Errorf("expected 0 items before load, got %d", pv.ItemCount())
	}

	_ = pv.Load(context.Background())
	if pv.ItemCount() != 1 {
		t.Errorf("expected 1 item after load, got %d", pv.ItemCount())
	}
}

func TestPoolsView_Loaded(t *testing.T) {
	mock := &truenas.MockDatasetService{
		ListPoolsFunc: func(ctx context.Context) ([]truenas.Pool, error) {
			return []truenas.Pool{}, nil
		},
	}

	pv := newPoolsView(mock)
	if pv.Loaded() {
		t.Error("expected Loaded()=false before Load()")
	}

	_ = pv.Load(context.Background())
	if !pv.Loaded() {
		t.Error("expected Loaded()=true after Load()")
	}
}

func TestPoolsView_Loaded_StaysFalseOnError(t *testing.T) {
	mock := &truenas.MockDatasetService{
		ListPoolsFunc: func(ctx context.Context) ([]truenas.Pool, error) {
			return nil, context.DeadlineExceeded
		},
	}

	pv := newPoolsView(mock)
	_ = pv.Load(context.Background())
	if pv.Loaded() {
		t.Error("expected Loaded()=false after failed Load()")
	}
}

func TestPoolsView_Stale(t *testing.T) {
	mock := &truenas.MockDatasetService{
		ListPoolsFunc: func(ctx context.Context) ([]truenas.Pool, error) {
			return []truenas.Pool{}, nil
		},
	}

	// Not loaded = stale
	pv := newPoolsView(mock)
	if !pv.Stale() {
		t.Error("expected Stale()=true before Load()")
	}

	// Just loaded = fresh
	_ = pv.Load(context.Background())
	if pv.Stale() {
		t.Error("expected Stale()=false immediately after Load()")
	}

	// With zero TTL = always stale after load
	pv2 := views.NewPoolsView(views.PoolsViewParams{Service: mock, StaleTTL: 0})
	_ = pv2.Load(context.Background())
	// time.Since(loadedAt) > 0 is true immediately
	time.Sleep(time.Millisecond)
	if !pv2.Stale() {
		t.Error("expected Stale()=true with zero TTL")
	}
}

func TestPoolsView_Draw_Loading(t *testing.T) {
	mock := &truenas.MockDatasetService{}
	pv := newPoolsView(mock)

	ctx := testDrawContext(80, 10)
	s, err := pv.Draw(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if s.Size.Width != 80 {
		t.Errorf("expected width=80, got %d", s.Size.Width)
	}
}

func TestPoolsView_Draw_WithData(t *testing.T) {
	mock := &truenas.MockDatasetService{
		ListPoolsFunc: func(ctx context.Context) ([]truenas.Pool, error) {
			return []truenas.Pool{
				{ID: 1, Name: "tank", Status: "ONLINE", Size: 1099511627776, Allocated: 549755813888, Free: 549755813888},
				{ID: 2, Name: "backup", Status: "DEGRADED", Size: 2199023255552, Allocated: 0, Free: 2199023255552},
			}, nil
		},
	}

	pv := newPoolsView(mock)
	_ = pv.Load(context.Background())

	ctx := testDrawContext(80, 10)
	s, err := pv.Draw(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if s.Size.Width != 80 {
		t.Errorf("expected surface width=80, got %d", s.Size.Width)
	}
	if s.Size.Height != 10 {
		t.Errorf("expected surface height=10, got %d", s.Size.Height)
	}
}

func TestPoolsView_Draw_Empty(t *testing.T) {
	mock := &truenas.MockDatasetService{
		ListPoolsFunc: func(ctx context.Context) ([]truenas.Pool, error) {
			return []truenas.Pool{}, nil
		},
	}

	pv := newPoolsView(mock)
	_ = pv.Load(context.Background())

	ctx := testDrawContext(80, 10)
	s, err := pv.Draw(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if s.Size.Width != 80 {
		t.Errorf("expected surface width=80, got %d", s.Size.Width)
	}
}

func TestPoolsView_HandleEvent(t *testing.T) {
	mock := &truenas.MockDatasetService{
		ListPoolsFunc: func(ctx context.Context) ([]truenas.Pool, error) {
			return []truenas.Pool{
				{ID: 1, Name: "tank", Status: "ONLINE"},
			}, nil
		},
	}

	pv := newPoolsView(mock)
	_ = pv.Load(context.Background())

	cmd, err := pv.HandleEvent(vaxis.Key{Keycode: 'j'}, vxfw.EventPhase(0))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	_ = cmd
}
