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

func newDatasetsView(mock *truenas.MockDatasetService) *views.DatasetsView {
	return views.NewDatasetsView(views.DatasetsViewParams{
		Service:  mock,
		StaleTTL: 30 * time.Second,
	})
}

func TestDatasetsView_Load(t *testing.T) {
	mock := &truenas.MockDatasetService{
		ListDatasetsFunc: func(ctx context.Context) ([]truenas.Dataset, error) {
			return []truenas.Dataset{
				{ID: "tank/data", Name: "data", Pool: "tank", Mountpoint: "/mnt/tank/data", Compression: "lz4", Used: 1073741824, Available: 549755813888},
				{ID: "tank/media", Name: "media", Pool: "tank", Mountpoint: "/mnt/tank/media", Compression: "off", Used: 0, Available: 549755813888},
			}, nil
		},
	}

	dv := newDatasetsView(mock)
	err := dv.Load(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	datasets := dv.Datasets()
	if len(datasets) != 2 {
		t.Fatalf("expected 2 datasets, got %d", len(datasets))
	}
}

func TestDatasetsView_Load_Error(t *testing.T) {
	mock := &truenas.MockDatasetService{
		ListDatasetsFunc: func(ctx context.Context) ([]truenas.Dataset, error) {
			return nil, context.DeadlineExceeded
		},
	}

	dv := newDatasetsView(mock)
	err := dv.Load(context.Background())
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestDatasetsView_ItemCount(t *testing.T) {
	mock := &truenas.MockDatasetService{
		ListDatasetsFunc: func(ctx context.Context) ([]truenas.Dataset, error) {
			return []truenas.Dataset{
				{ID: "tank/data", Name: "data", Pool: "tank"},
			}, nil
		},
	}

	dv := newDatasetsView(mock)
	if dv.ItemCount() != 0 {
		t.Errorf("expected 0 items before load, got %d", dv.ItemCount())
	}
	_ = dv.Load(context.Background())
	if dv.ItemCount() != 1 {
		t.Errorf("expected 1 item after load, got %d", dv.ItemCount())
	}
}

func TestDatasetsView_Loaded(t *testing.T) {
	mock := &truenas.MockDatasetService{
		ListDatasetsFunc: func(ctx context.Context) ([]truenas.Dataset, error) {
			return []truenas.Dataset{}, nil
		},
	}

	dv := newDatasetsView(mock)
	if dv.Loaded() {
		t.Error("expected Loaded()=false before Load()")
	}

	_ = dv.Load(context.Background())
	if !dv.Loaded() {
		t.Error("expected Loaded()=true after Load()")
	}
}

func TestDatasetsView_Stale(t *testing.T) {
	mock := &truenas.MockDatasetService{
		ListDatasetsFunc: func(ctx context.Context) ([]truenas.Dataset, error) {
			return []truenas.Dataset{}, nil
		},
	}

	dv := newDatasetsView(mock)
	if !dv.Stale() {
		t.Error("expected Stale()=true before Load()")
	}

	_ = dv.Load(context.Background())
	if dv.Stale() {
		t.Error("expected Stale()=false immediately after Load()")
	}
}

func TestDatasetsView_Draw_Loading(t *testing.T) {
	mock := &truenas.MockDatasetService{}
	dv := newDatasetsView(mock)

	ctx := testDrawContext(80, 10)
	s, err := dv.Draw(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if s.Size.Width != 80 {
		t.Errorf("expected width=80, got %d", s.Size.Width)
	}
}

func TestDatasetsView_Draw_WithData(t *testing.T) {
	mock := &truenas.MockDatasetService{
		ListDatasetsFunc: func(ctx context.Context) ([]truenas.Dataset, error) {
			return []truenas.Dataset{
				{ID: "tank/data", Name: "data", Pool: "tank", Mountpoint: "/mnt/tank/data", Compression: "lz4", Used: 1073741824, Available: 549755813888},
				{ID: "tank/media", Name: "media", Pool: "tank", Mountpoint: "/mnt/tank/media", Compression: "off", Used: 0, Available: 549755813888},
			}, nil
		},
	}

	dv := newDatasetsView(mock)
	_ = dv.Load(context.Background())

	ctx := testDrawContext(100, 10)
	s, err := dv.Draw(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if s.Size.Width != 100 {
		t.Errorf("expected surface width=100, got %d", s.Size.Width)
	}
	if s.Size.Height != 10 {
		t.Errorf("expected surface height=10, got %d", s.Size.Height)
	}
}

func TestDatasetsView_Draw_Empty(t *testing.T) {
	mock := &truenas.MockDatasetService{
		ListDatasetsFunc: func(ctx context.Context) ([]truenas.Dataset, error) {
			return []truenas.Dataset{}, nil
		},
	}

	dv := newDatasetsView(mock)
	_ = dv.Load(context.Background())

	ctx := testDrawContext(80, 10)
	s, err := dv.Draw(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if s.Size.Width != 80 {
		t.Errorf("expected surface width=80, got %d", s.Size.Width)
	}
}

func TestDatasetsView_HandleEvent(t *testing.T) {
	mock := &truenas.MockDatasetService{
		ListDatasetsFunc: func(ctx context.Context) ([]truenas.Dataset, error) {
			return []truenas.Dataset{
				{ID: "tank/data", Name: "data", Pool: "tank"},
			}, nil
		},
	}

	dv := newDatasetsView(mock)
	_ = dv.Load(context.Background())

	cmd, err := dv.HandleEvent(vaxis.Key{Keycode: 'j'}, vxfw.EventPhase(0))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	_ = cmd
}
