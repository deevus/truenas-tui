package app

import (
	"context"
	"fmt"
	"log"
	"time"

	"git.sr.ht/~rockorager/vaxis"
	"git.sr.ht/~rockorager/vaxis/vxfw"
	"git.sr.ht/~rockorager/vaxis/vxfw/richtext"
	"github.com/deevus/truenas-tui/internal"
	"github.com/deevus/truenas-tui/views"
	"github.com/deevus/truenas-tui/widgets"
)

// Connected is posted when the background connection goroutine succeeds.
type Connected struct {
	Services *internal.Services
}

// ConnectFailed is posted when the background connection goroutine fails.
type ConnectFailed struct {
	Err error
}

// Params holds configuration for creating an App.
type Params struct {
	ServerName string
	StaleTTL   time.Duration
	Services   *internal.Services                                   // immediate (tests)
	Connect    func(ctx context.Context) (*internal.Services, error) // async (main)
}

// App is the root vxfw widget for truenas-tui.
type App struct {
	serverName string
	staleTTL   time.Duration
	tabBar     *widgets.TabBar
	services   *internal.Services
	dashboard  *views.DashboardView
	pools      *views.PoolsView
	datasets   *views.DatasetsView
	snapshots  *views.SnapshotsView
	postEvent  func(vaxis.Event)
	connectFn  func(ctx context.Context) (*internal.Services, error)
	connected  bool
	connectErr error
}

// New creates the root App widget.
// If p.Services is set, the app starts connected immediately (useful for tests).
// If p.Connect is set, the app starts in a connecting state and runs the
// callback from the Init event in a background goroutine.
func New(p Params) *App {
	a := &App{
		serverName: p.ServerName,
		staleTTL:   p.StaleTTL,
		connectFn:  p.Connect,
		tabBar:     widgets.NewTabBar([]string{"Dashboard", "Pools", "Datasets", "Snapshots"}),
	}
	if p.Services != nil {
		a.initServices(p.Services)
	}
	return a
}

// initServices creates views backed by the given services and marks the app
// as connected.
func (a *App) initServices(svc *internal.Services) {
	a.services = svc
	a.dashboard = views.NewDashboardView(views.DashboardViewParams{
		System:     svc.System,
		Reporting:  svc.Reporting,
		Interfaces: svc.Interfaces,
		Apps:       svc.Apps,
		PostEvent:  a.postEvent,
	})
	a.pools = views.NewPoolsView(views.PoolsViewParams{Service: svc.Datasets, StaleTTL: a.staleTTL})
	a.datasets = views.NewDatasetsView(views.DatasetsViewParams{Service: svc.Datasets, StaleTTL: a.staleTTL})
	a.snapshots = views.NewSnapshotsView(views.SnapshotsViewParams{Service: svc.Snapshots, StaleTTL: a.staleTTL})
	a.connected = true
}

// SetPostEvent sets the function used to post events to the vaxis event loop.
func (a *App) SetPostEvent(fn func(vaxis.Event)) {
	a.postEvent = fn
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

// Connected reports whether the app has an active connection.
func (a *App) IsConnected() bool {
	return a.connected
}

// LoadAll loads data for all views in parallel using goroutines.
// Each view posts a ViewLoaded event when done.
func (a *App) LoadAll(ctx context.Context) {
	if !a.connected {
		return
	}
	for tab := 0; tab < 4; tab++ {
		go func(t int) {
			var err error
			switch t {
			case 0:
				err = a.dashboard.Load(ctx)
			case 1:
				err = a.pools.Load(ctx)
			case 2:
				err = a.datasets.Load(ctx)
			case 3:
				err = a.snapshots.Load(ctx)
			}
			if a.postEvent != nil {
				a.postEvent(views.ViewLoaded{Tab: t, Err: err})
			}
		}(tab)
	}
}

// LoadActiveView fetches data for the currently active view.
func (a *App) LoadActiveView(ctx context.Context) error {
	if !a.connected {
		return nil
	}
	switch a.tabBar.Active() {
	case 0:
		return a.dashboard.Load(ctx)
	case 1:
		return a.pools.Load(ctx)
	case 2:
		return a.datasets.Load(ctx)
	case 3:
		return a.snapshots.Load(ctx)
	}
	return nil
}

func (a *App) activeView() vxfw.Widget {
	if !a.connected {
		return nil
	}
	switch a.tabBar.Active() {
	case 0:
		return a.dashboard
	case 1:
		return a.pools
	case 2:
		return a.datasets
	case 3:
		return a.snapshots
	default:
		return a.dashboard
	}
}

// drawMessage renders a single dimmed text message.
func drawMessage(ctx vxfw.DrawContext, owner vxfw.Widget, text string) (vxfw.Surface, error) {
	s := vxfw.NewSurface(ctx.Max.Width, ctx.Max.Height, owner)
	label := richtext.New([]vaxis.Segment{
		{Text: text, Style: vaxis.Style{Attribute: vaxis.AttrDim}},
	})
	labelSurf, err := label.Draw(ctx.WithMax(vxfw.Size{Width: ctx.Max.Width, Height: 1}))
	if err != nil {
		return vxfw.Surface{}, err
	}
	s.AddChild(0, 0, labelSurf)
	return s, nil
}

// Draw renders the tab bar and active view, or a status message if not connected.
func (a *App) Draw(ctx vxfw.DrawContext) (vxfw.Surface, error) {
	if a.connectErr != nil {
		return drawMessage(ctx, a, fmt.Sprintf("Connection failed: %v", a.connectErr))
	}
	if !a.connected {
		return drawMessage(ctx, a, fmt.Sprintf("Connecting to %s...", a.serverName))
	}

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
		if ev.Matches('q') {
			return vxfw.QuitCmd{}, nil
		}
		if !a.connected {
			return nil, nil
		}
		prev := a.tabBar.Active()
		switch {
		case ev.Matches('r'):
			a.loadActiveViewAsync()
			return vxfw.ConsumeAndRedraw(), nil
		case ev.Matches('1'):
			a.tabBar.SetActive(0)
		case ev.Matches('2'):
			a.tabBar.SetActive(1)
		case ev.Matches('3'):
			a.tabBar.SetActive(2)
		case ev.Matches('4'):
			a.tabBar.SetActive(3)
		case ev.Matches(vaxis.KeyTab):
			a.tabBar.Next()
		case ev.Matches(vaxis.KeyTab, vaxis.ModShift):
			a.tabBar.Prev()
		default:
			return nil, nil
		}
		if a.tabBar.Active() != prev {
			a.refetchIfStale()
		}
		return vxfw.ConsumeAndRedraw(), nil
	}
	return nil, nil
}

// loadActiveViewAsync loads the active view's data in a background goroutine.
func (a *App) loadActiveViewAsync() {
	if !a.connected {
		return
	}
	tab := a.tabBar.Active()
	go func() {
		err := a.LoadActiveView(context.Background())
		if a.postEvent != nil {
			a.postEvent(views.ViewLoaded{Tab: tab, Err: err})
		}
	}()
}

// refetchIfStale reloads the active view's data in the background if stale.
func (a *App) refetchIfStale() {
	if !a.connected {
		return
	}
	var stale bool
	switch a.tabBar.Active() {
	case 0:
		// Dashboard is streaming, never stale
		return
	case 1:
		stale = a.pools.Stale()
	case 2:
		stale = a.datasets.Stale()
	case 3:
		stale = a.snapshots.Stale()
	}
	if stale {
		a.loadActiveViewAsync()
	}
}

// HandleEvent delegates to the active view, and handles custom events.
func (a *App) HandleEvent(ev vaxis.Event, phase vxfw.EventPhase) (vxfw.Command, error) {
	switch ev := ev.(type) {
	case vxfw.Init:
		if a.connectFn != nil {
			go func() {
				svc, err := a.connectFn(context.Background())
				if err != nil {
					a.postEvent(ConnectFailed{Err: err})
				} else {
					a.postEvent(Connected{Services: svc})
				}
			}()
		}
		return nil, nil
	case Connected:
		a.initServices(ev.Services)
		a.LoadAll(context.Background())
		return vxfw.RedrawCmd{}, nil
	case ConnectFailed:
		a.connectErr = ev.Err
		return vxfw.RedrawCmd{}, nil
	case views.ViewLoaded:
		if ev.Err != nil {
			log.Printf("error loading tab %d: %v", ev.Tab, ev.Err)
		}
		// Start dashboard subscriptions once it has loaded
		if ev.Tab == 0 && ev.Err == nil && a.dashboard != nil {
			a.dashboard.StartSubscriptions(context.Background())
		}
		return vxfw.RedrawCmd{}, nil
	case views.DashboardUpdated:
		return vxfw.RedrawCmd{}, nil
	default:
		type handler interface {
			HandleEvent(vaxis.Event, vxfw.EventPhase) (vxfw.Command, error)
		}
		if v := a.activeView(); v != nil {
			if h, ok := v.(handler); ok {
				return h.HandleEvent(ev, phase)
			}
		}
	}
	return nil, nil
}
