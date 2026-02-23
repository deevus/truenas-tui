package views_test

import (
	"context"
	"testing"

	"github.com/deevus/truenas-go"
	"github.com/deevus/truenas-tui/views"
)

func TestDatasetsView_Load(t *testing.T) {
	mock := &truenas.MockDatasetService{
		ListDatasetsFunc: func(ctx context.Context) ([]truenas.Dataset, error) {
			return []truenas.Dataset{
				{ID: "tank/data", Name: "data", Pool: "tank", Mountpoint: "/mnt/tank/data", Compression: "lz4", Used: 1073741824, Available: 549755813888},
				{ID: "tank/media", Name: "media", Pool: "tank", Mountpoint: "/mnt/tank/media", Compression: "off", Used: 0, Available: 549755813888},
			}, nil
		},
	}

	dv := views.NewDatasetsView(mock)
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

	dv := views.NewDatasetsView(mock)
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

	dv := views.NewDatasetsView(mock)
	if dv.ItemCount() != 0 {
		t.Errorf("expected 0 items before load, got %d", dv.ItemCount())
	}
	_ = dv.Load(context.Background())
	if dv.ItemCount() != 1 {
		t.Errorf("expected 1 item after load, got %d", dv.ItemCount())
	}
}
