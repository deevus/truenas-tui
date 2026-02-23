package views_test

import (
	"context"
	"testing"

	"github.com/deevus/truenas-go"
	"github.com/deevus/truenas-tui/views"
)

func TestSnapshotsView_Load(t *testing.T) {
	mock := &truenas.MockSnapshotService{
		ListFunc: func(ctx context.Context) ([]truenas.Snapshot, error) {
			return []truenas.Snapshot{
				{ID: "tank/data@auto-2024-01-01", Dataset: "tank/data", SnapshotName: "auto-2024-01-01", Used: 1024, Referenced: 1073741824},
				{ID: "tank/data@manual", Dataset: "tank/data", SnapshotName: "manual", Used: 0, Referenced: 1073741824, HasHold: true},
			}, nil
		},
	}

	sv := views.NewSnapshotsView(mock)
	err := sv.Load(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	snaps := sv.Snapshots()
	if len(snaps) != 2 {
		t.Fatalf("expected 2 snapshots, got %d", len(snaps))
	}
}

func TestSnapshotsView_Load_Error(t *testing.T) {
	mock := &truenas.MockSnapshotService{
		ListFunc: func(ctx context.Context) ([]truenas.Snapshot, error) {
			return nil, context.DeadlineExceeded
		},
	}

	sv := views.NewSnapshotsView(mock)
	err := sv.Load(context.Background())
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestSnapshotsView_ItemCount(t *testing.T) {
	mock := &truenas.MockSnapshotService{
		ListFunc: func(ctx context.Context) ([]truenas.Snapshot, error) {
			return []truenas.Snapshot{
				{ID: "tank/data@snap1", Dataset: "tank/data", SnapshotName: "snap1"},
			}, nil
		},
	}

	sv := views.NewSnapshotsView(mock)
	if sv.ItemCount() != 0 {
		t.Errorf("expected 0 before load, got %d", sv.ItemCount())
	}
	_ = sv.Load(context.Background())
	if sv.ItemCount() != 1 {
		t.Errorf("expected 1 after load, got %d", sv.ItemCount())
	}
}

func TestSnapshotsView_SelectedSnapshot(t *testing.T) {
	mock := &truenas.MockSnapshotService{
		ListFunc: func(ctx context.Context) ([]truenas.Snapshot, error) {
			return []truenas.Snapshot{
				{ID: "tank/data@snap1", Dataset: "tank/data", SnapshotName: "snap1"},
				{ID: "tank/data@snap2", Dataset: "tank/data", SnapshotName: "snap2"},
			}, nil
		},
	}

	sv := views.NewSnapshotsView(mock)
	_ = sv.Load(context.Background())

	snap := sv.SelectedSnapshot()
	if snap == nil {
		t.Fatal("expected selected snapshot")
	}
	if snap.SnapshotName != "snap1" {
		t.Errorf("expected snap1, got %s", snap.SnapshotName)
	}
}
