package internal_test

import (
	"testing"

	"github.com/deevus/truenas-go"
	"github.com/deevus/truenas-tui/internal"
)

func TestNewServices(t *testing.T) {
	svc := internal.NewServices(
		&truenas.MockDatasetService{},
		&truenas.MockSnapshotService{},
		&truenas.MockSystemService{},
		&truenas.MockReportingService{},
		&truenas.MockInterfaceService{},
		&truenas.MockAppService{},
	)

	if svc.Datasets == nil {
		t.Fatal("expected Datasets service")
	}
	if svc.Snapshots == nil {
		t.Fatal("expected Snapshots service")
	}
	if svc.System == nil {
		t.Fatal("expected System service")
	}
	if svc.Reporting == nil {
		t.Fatal("expected Reporting service")
	}
	if svc.Interfaces == nil {
		t.Fatal("expected Interfaces service")
	}
	if svc.Apps == nil {
		t.Fatal("expected Apps service")
	}
}
