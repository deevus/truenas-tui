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

func newSnapshotsView(mock *truenas.MockSnapshotService) *views.SnapshotsView {
	return views.NewSnapshotsView(views.SnapshotsViewParams{
		Service:  mock,
		StaleTTL: 30 * time.Second,
	})
}

func TestSnapshotsView_Load(t *testing.T) {
	mock := &truenas.MockSnapshotService{
		ListFunc: func(ctx context.Context) ([]truenas.Snapshot, error) {
			return []truenas.Snapshot{
				{ID: "tank/data@auto-2024-01-01", Dataset: "tank/data", SnapshotName: "auto-2024-01-01", Used: 1024, Referenced: 1073741824},
				{ID: "tank/data@manual", Dataset: "tank/data", SnapshotName: "manual", Used: 0, Referenced: 1073741824, HasHold: true},
			}, nil
		},
	}

	sv := newSnapshotsView(mock)
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

	sv := newSnapshotsView(mock)
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

	sv := newSnapshotsView(mock)
	if sv.ItemCount() != 0 {
		t.Errorf("expected 0 before load, got %d", sv.ItemCount())
	}
	_ = sv.Load(context.Background())
	if sv.ItemCount() != 1 {
		t.Errorf("expected 1 after load, got %d", sv.ItemCount())
	}
}

func TestSnapshotsView_Loaded(t *testing.T) {
	mock := &truenas.MockSnapshotService{
		ListFunc: func(ctx context.Context) ([]truenas.Snapshot, error) {
			return []truenas.Snapshot{}, nil
		},
	}

	sv := newSnapshotsView(mock)
	if sv.Loaded() {
		t.Error("expected Loaded()=false before Load()")
	}

	_ = sv.Load(context.Background())
	if !sv.Loaded() {
		t.Error("expected Loaded()=true after Load()")
	}
}

func TestSnapshotsView_Stale(t *testing.T) {
	mock := &truenas.MockSnapshotService{
		ListFunc: func(ctx context.Context) ([]truenas.Snapshot, error) {
			return []truenas.Snapshot{}, nil
		},
	}

	sv := newSnapshotsView(mock)
	if !sv.Stale() {
		t.Error("expected Stale()=true before Load()")
	}

	_ = sv.Load(context.Background())
	if sv.Stale() {
		t.Error("expected Stale()=false immediately after Load()")
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

	sv := newSnapshotsView(mock)
	_ = sv.Load(context.Background())

	snap := sv.SelectedSnapshot()
	if snap == nil {
		t.Fatal("expected selected snapshot")
	}
	if snap.SnapshotName != "snap1" {
		t.Errorf("expected snap1, got %s", snap.SnapshotName)
	}
}

func TestSnapshotsView_SelectedSnapshot_Empty(t *testing.T) {
	mock := &truenas.MockSnapshotService{
		ListFunc: func(ctx context.Context) ([]truenas.Snapshot, error) {
			return []truenas.Snapshot{}, nil
		},
	}

	sv := newSnapshotsView(mock)
	_ = sv.Load(context.Background())

	snap := sv.SelectedSnapshot()
	if snap != nil {
		t.Errorf("expected nil when no snapshots loaded, got %v", snap)
	}
}

func TestSnapshotsView_SelectedSnapshot_BeforeLoad(t *testing.T) {
	mock := &truenas.MockSnapshotService{}
	sv := newSnapshotsView(mock)

	snap := sv.SelectedSnapshot()
	if snap != nil {
		t.Errorf("expected nil before loading, got %v", snap)
	}
}

func TestSnapshotsView_Draw_Loading(t *testing.T) {
	mock := &truenas.MockSnapshotService{}
	sv := newSnapshotsView(mock)

	ctx := testDrawContext(80, 10)
	s, err := sv.Draw(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if s.Size.Width != 80 {
		t.Errorf("expected width=80, got %d", s.Size.Width)
	}
}

func TestSnapshotsView_Draw_WithData(t *testing.T) {
	mock := &truenas.MockSnapshotService{
		ListFunc: func(ctx context.Context) ([]truenas.Snapshot, error) {
			return []truenas.Snapshot{
				{ID: "tank/data@auto-2024-01-01", Dataset: "tank/data", SnapshotName: "auto-2024-01-01", Used: 1024, Referenced: 1073741824},
				{ID: "tank/data@manual", Dataset: "tank/data", SnapshotName: "manual", Used: 0, Referenced: 1073741824, HasHold: true},
			}, nil
		},
	}

	sv := newSnapshotsView(mock)
	_ = sv.Load(context.Background())

	ctx := testDrawContext(100, 10)
	s, err := sv.Draw(ctx)
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

func TestSnapshotsView_Draw_Empty(t *testing.T) {
	mock := &truenas.MockSnapshotService{
		ListFunc: func(ctx context.Context) ([]truenas.Snapshot, error) {
			return []truenas.Snapshot{}, nil
		},
	}

	sv := newSnapshotsView(mock)
	_ = sv.Load(context.Background())

	ctx := testDrawContext(80, 10)
	s, err := sv.Draw(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if s.Size.Width != 80 {
		t.Errorf("expected surface width=80, got %d", s.Size.Width)
	}
}

func TestSnapshotsView_HandleEvent(t *testing.T) {
	mock := &truenas.MockSnapshotService{
		ListFunc: func(ctx context.Context) ([]truenas.Snapshot, error) {
			return []truenas.Snapshot{
				{ID: "tank/data@snap1", Dataset: "tank/data", SnapshotName: "snap1"},
			}, nil
		},
	}

	sv := newSnapshotsView(mock)
	_ = sv.Load(context.Background())

	cmd, err := sv.HandleEvent(vaxis.Key{Keycode: 'j'}, vxfw.EventPhase(0))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	_ = cmd
}
