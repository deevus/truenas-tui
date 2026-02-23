package views_test

import (
	"context"
	"testing"

	"git.sr.ht/~rockorager/vaxis"
	"git.sr.ht/~rockorager/vaxis/vxfw"
	"github.com/deevus/truenas-go"
	"github.com/deevus/truenas-tui/views"
)

func TestPoolsView_Load(t *testing.T) {
	mock := &truenas.MockDatasetService{
		ListPoolsFunc: func(ctx context.Context) ([]truenas.Pool, error) {
			return []truenas.Pool{
				{ID: 1, Name: "tank", Status: "ONLINE", Size: 1099511627776, Allocated: 549755813888, Free: 549755813888},
				{ID: 2, Name: "backup", Status: "ONLINE", Size: 2199023255552, Allocated: 0, Free: 2199023255552},
			}, nil
		},
	}

	pv := views.NewPoolsView(mock)
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

	pv := views.NewPoolsView(mock)
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

	pv := views.NewPoolsView(mock)
	if pv.ItemCount() != 0 {
		t.Errorf("expected 0 items before load, got %d", pv.ItemCount())
	}

	_ = pv.Load(context.Background())
	if pv.ItemCount() != 1 {
		t.Errorf("expected 1 item after load, got %d", pv.ItemCount())
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

	pv := views.NewPoolsView(mock)
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

	pv := views.NewPoolsView(mock)
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

	pv := views.NewPoolsView(mock)
	_ = pv.Load(context.Background())

	// HandleEvent should not panic for a basic key event
	cmd, err := pv.HandleEvent(vaxis.Key{Keycode: 'j'}, vxfw.EventPhase(0))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// cmd may or may not be nil depending on the list widget
	_ = cmd
}
