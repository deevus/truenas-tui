package internal_test

import (
	"testing"

	"github.com/deevus/truenas-go"
	"github.com/deevus/truenas-tui/internal"
)

func TestNewServices(t *testing.T) {
	mock := &truenas.MockDatasetService{}
	mockSnap := &truenas.MockSnapshotService{}

	svc := internal.NewServices(mock, mockSnap)

	if svc.Datasets == nil {
		t.Fatal("expected Datasets service")
	}
	if svc.Snapshots == nil {
		t.Fatal("expected Snapshots service")
	}
}
