package app_test

import (
	"testing"

	"github.com/deevus/truenas-go"
	"github.com/deevus/truenas-tui/app"
	"github.com/deevus/truenas-tui/internal"
)

func newTestServices() *internal.Services {
	return internal.NewServices(
		&truenas.MockDatasetService{},
		&truenas.MockSnapshotService{},
	)
}

func TestApp_New(t *testing.T) {
	a := app.New(newTestServices(), "test-server")
	if a == nil {
		t.Fatal("expected non-nil app")
	}
}

func TestApp_ActiveTab(t *testing.T) {
	a := app.New(newTestServices(), "test-server")
	if a.ActiveTab() != 0 {
		t.Errorf("expected initial tab 0, got %d", a.ActiveTab())
	}
}

func TestApp_SetTab(t *testing.T) {
	a := app.New(newTestServices(), "test-server")
	a.SetTab(1)
	if a.ActiveTab() != 1 {
		t.Errorf("expected tab 1, got %d", a.ActiveTab())
	}
	a.SetTab(2)
	if a.ActiveTab() != 2 {
		t.Errorf("expected tab 2, got %d", a.ActiveTab())
	}
}

func TestApp_ServerName(t *testing.T) {
	a := app.New(newTestServices(), "home")
	if a.ServerName() != "home" {
		t.Errorf("expected server name home, got %s", a.ServerName())
	}
}
