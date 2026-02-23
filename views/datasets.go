package views

import (
	"context"
	"fmt"

	"git.sr.ht/~rockorager/vaxis"
	"git.sr.ht/~rockorager/vaxis/vxfw"
	"git.sr.ht/~rockorager/vaxis/vxfw/list"
	"git.sr.ht/~rockorager/vaxis/vxfw/richtext"
	"github.com/deevus/truenas-go"
	"github.com/dustin/go-humanize"
)

// DatasetsView displays a list of TrueNAS datasets.
type DatasetsView struct {
	service  truenas.DatasetServiceAPI
	datasets []truenas.Dataset
	list     list.Dynamic
}

// NewDatasetsView creates a DatasetsView backed by the given service.
func NewDatasetsView(svc truenas.DatasetServiceAPI) *DatasetsView {
	dv := &DatasetsView{service: svc}
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
	return nil
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

// Draw renders the datasets list.
func (dv *DatasetsView) Draw(ctx vxfw.DrawContext) (vxfw.Surface, error) {
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
