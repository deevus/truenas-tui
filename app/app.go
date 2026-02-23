package app

import (
	"context"

	"git.sr.ht/~rockorager/vaxis"
	"git.sr.ht/~rockorager/vaxis/vxfw"
	"github.com/deevus/truenas-tui/internal"
	"github.com/deevus/truenas-tui/views"
	"github.com/deevus/truenas-tui/widgets"
)

// App is the root vxfw widget for truenas-tui.
type App struct {
	services   *internal.Services
	serverName string
	tabBar     *widgets.TabBar
	pools      *views.PoolsView
	datasets   *views.DatasetsView
	snapshots  *views.SnapshotsView
}

// New creates the root App widget connected to the given services.
func New(svc *internal.Services, serverName string) *App {
	return &App{
		services:   svc,
		serverName: serverName,
		tabBar:     widgets.NewTabBar([]string{"Pools", "Datasets", "Snapshots"}),
		pools:      views.NewPoolsView(svc.Datasets),
		datasets:   views.NewDatasetsView(svc.Datasets),
		snapshots:  views.NewSnapshotsView(svc.Snapshots),
	}
}

// ActiveTab returns the current tab index.
func (a *App) ActiveTab() int {
	return a.tabBar.Active()
}

// SetTab switches to the given tab index.
func (a *App) SetTab(i int) {
	a.tabBar.SetActive(i)
}

// ServerName returns the connected server profile name.
func (a *App) ServerName() string {
	return a.serverName
}

// LoadActiveView fetches data for the currently active view.
func (a *App) LoadActiveView(ctx context.Context) error {
	switch a.tabBar.Active() {
	case 0:
		return a.pools.Load(ctx)
	case 1:
		return a.datasets.Load(ctx)
	case 2:
		return a.snapshots.Load(ctx)
	}
	return nil
}

func (a *App) activeView() vxfw.Widget {
	switch a.tabBar.Active() {
	case 0:
		return a.pools
	case 1:
		return a.datasets
	case 2:
		return a.snapshots
	default:
		return a.pools
	}
}

// Draw renders the tab bar and active view.
func (a *App) Draw(ctx vxfw.DrawContext) (vxfw.Surface, error) {
	s := vxfw.NewSurface(ctx.Max.Width, ctx.Max.Height, a)

	// Tab bar (1 row)
	tabCtx := ctx.WithMax(vxfw.Size{Width: ctx.Max.Width, Height: 1})
	tabSurf, err := a.tabBar.Draw(tabCtx)
	if err != nil {
		return vxfw.Surface{}, err
	}
	s.AddChild(0, 0, tabSurf)

	// Active view (remaining space)
	viewCtx := ctx.WithMax(vxfw.Size{Width: ctx.Max.Width, Height: ctx.Max.Height - 1})
	viewSurf, err := a.activeView().Draw(viewCtx)
	if err != nil {
		return vxfw.Surface{}, err
	}
	s.AddChild(0, 1, viewSurf)

	return s, nil
}

// CaptureEvent handles global keybindings before views process them.
func (a *App) CaptureEvent(ev vaxis.Event) (vxfw.Command, error) {
	switch ev := ev.(type) {
	case vaxis.Key:
		switch {
		case ev.Matches('q'):
			return vxfw.QuitCmd{}, nil
		case ev.Matches('1'):
			a.tabBar.SetActive(0)
			return vxfw.ConsumeAndRedraw(), nil
		case ev.Matches('2'):
			a.tabBar.SetActive(1)
			return vxfw.ConsumeAndRedraw(), nil
		case ev.Matches('3'):
			a.tabBar.SetActive(2)
			return vxfw.ConsumeAndRedraw(), nil
		case ev.Matches(vaxis.KeyTab):
			a.tabBar.Next()
			return vxfw.ConsumeAndRedraw(), nil
		case ev.Matches(vaxis.KeyTab, vaxis.ModShift):
			a.tabBar.Prev()
			return vxfw.ConsumeAndRedraw(), nil
		}
	}
	return nil, nil
}

// HandleEvent delegates to the active view.
func (a *App) HandleEvent(ev vaxis.Event, phase vxfw.EventPhase) (vxfw.Command, error) {
	type handler interface {
		HandleEvent(vaxis.Event, vxfw.EventPhase) (vxfw.Command, error)
	}
	if h, ok := a.activeView().(handler); ok {
		return h.HandleEvent(ev, phase)
	}
	return nil, nil
}
