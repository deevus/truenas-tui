package views

import (
	"context"
	"fmt"
	"time"

	"git.sr.ht/~rockorager/vaxis"
	"git.sr.ht/~rockorager/vaxis/vxfw"
	"git.sr.ht/~rockorager/vaxis/vxfw/list"
	"git.sr.ht/~rockorager/vaxis/vxfw/richtext"
	"github.com/deevus/truenas-go"
	"github.com/dustin/go-humanize"
)

// DatasetsViewParams holds configuration for creating a DatasetsView.
type DatasetsViewParams struct {
	Service  truenas.DatasetServiceAPI
	StaleTTL time.Duration
}

// DatasetsView displays a list of TrueNAS datasets.
type DatasetsView struct {
	service  truenas.DatasetServiceAPI
	datasets []truenas.Dataset
	list     list.Dynamic
	loaded   bool
	loadedAt time.Time
	staleTTL time.Duration
}

// NewDatasetsView creates a DatasetsView backed by the given params.
func NewDatasetsView(p DatasetsViewParams) *DatasetsView {
	dv := &DatasetsView{
		service:  p.Service,
		staleTTL: p.StaleTTL,
	}
	dv.list.DrawCursor = true
	dv.list.Builder = dv.buildItem
	return dv
}

// Load fetches datasets from the service.
func (dv *DatasetsView) Load(ctx context.Context) error {
	datasets, err := dv.service.ListDatasets(ctx)
	if err != nil {
		return err
	}
	dv.datasets = datasets
	dv.loaded = true
	dv.loadedAt = time.Now()
	return nil
}

// Loaded reports whether data has been successfully fetched.
func (dv *DatasetsView) Loaded() bool {
	return dv.loaded
}

// Stale reports whether the cached data is older than the configured TTL.
func (dv *DatasetsView) Stale() bool {
	if !dv.loaded {
		return true
	}
	return time.Since(dv.loadedAt) > dv.staleTTL
}

// Datasets returns the currently loaded datasets.
func (dv *DatasetsView) Datasets() []truenas.Dataset {
	return dv.datasets
}

// ItemCount returns the number of loaded datasets.
func (dv *DatasetsView) ItemCount() int {
	return len(dv.datasets)
}

func (dv *DatasetsView) buildItem(i uint, cursor uint) vxfw.Widget {
	if int(i) >= len(dv.datasets) {
		return nil
	}
	d := dv.datasets[i]

	return richtext.New([]vaxis.Segment{
		{Text: fmt.Sprintf("%-30s", d.ID)},
		{Text: fmt.Sprintf("%-10s", d.Compression)},
		{Text: fmt.Sprintf("%10s", humanize.IBytes(uint64(d.Used)))},
		{Text: fmt.Sprintf("%10s", humanize.IBytes(uint64(d.Available)))},
		{Text: fmt.Sprintf("  %s", d.Mountpoint)},
	})
}

// Draw renders the datasets list, or a loading state if data hasn't arrived.
func (dv *DatasetsView) Draw(ctx vxfw.DrawContext) (vxfw.Surface, error) {
	if !dv.loaded {
		return drawLoadingState(ctx, dv)
	}

	s := vxfw.NewSurface(ctx.Max.Width, ctx.Max.Height, dv)

	header := richtext.New([]vaxis.Segment{
		{Text: fmt.Sprintf("%-30s%-10s%10s%10s  %s", "NAME", "COMPRESS", "USED", "AVAIL", "MOUNTPOINT"),
			Style: vaxis.Style{Attribute: vaxis.AttrBold}},
	})
	headerSurf, err := header.Draw(ctx.WithMax(vxfw.Size{Width: ctx.Max.Width, Height: 1}))
	if err != nil {
		return vxfw.Surface{}, err
	}
	s.AddChild(0, 0, headerSurf)

	listCtx := ctx.WithMax(vxfw.Size{Width: ctx.Max.Width, Height: ctx.Max.Height - 1})
	listSurf, err := dv.list.Draw(listCtx)
	if err != nil {
		return vxfw.Surface{}, err
	}
	s.AddChild(0, 1, listSurf)

	return s, nil
}

// HandleEvent delegates to the list widget for navigation.
func (dv *DatasetsView) HandleEvent(ev vaxis.Event, phase vxfw.EventPhase) (vxfw.Command, error) {
	return dv.list.HandleEvent(ev, phase)
}
