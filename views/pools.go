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

// PoolsView displays a list of TrueNAS storage pools.
type PoolsView struct {
	service truenas.DatasetServiceAPI
	pools   []truenas.Pool
	list    list.Dynamic
}

// NewPoolsView creates a PoolsView backed by the given service.
func NewPoolsView(svc truenas.DatasetServiceAPI) *PoolsView {
	pv := &PoolsView{service: svc}
	pv.list.DrawCursor = true
	pv.list.Builder = pv.buildItem
	return pv
}

// Load fetches pools from the service.
func (pv *PoolsView) Load(ctx context.Context) error {
	pools, err := pv.service.ListPools(ctx)
	if err != nil {
		return err
	}
	pv.pools = pools
	return nil
}

// Pools returns the currently loaded pools.
func (pv *PoolsView) Pools() []truenas.Pool {
	return pv.pools
}

// ItemCount returns the number of loaded pools.
func (pv *PoolsView) ItemCount() int {
	return len(pv.pools)
}

func (pv *PoolsView) buildItem(i uint, cursor uint) vxfw.Widget {
	if int(i) >= len(pv.pools) {
		return nil
	}
	p := pv.pools[i]

	statusStyle := vaxis.Style{Foreground: vaxis.IndexColor(2)} // green
	if p.Status != "ONLINE" {
		statusStyle.Foreground = vaxis.IndexColor(1) // red
	}

	return richtext.New([]vaxis.Segment{
		{Text: fmt.Sprintf("%-20s", p.Name)},
		{Text: fmt.Sprintf("%-10s", p.Status), Style: statusStyle},
		{Text: fmt.Sprintf("%10s", humanize.IBytes(uint64(p.Size)))},
		{Text: fmt.Sprintf("%10s", humanize.IBytes(uint64(p.Allocated)))},
		{Text: fmt.Sprintf("%10s", humanize.IBytes(uint64(p.Free)))},
	})
}

// Draw renders the pools list.
func (pv *PoolsView) Draw(ctx vxfw.DrawContext) (vxfw.Surface, error) {
	s := vxfw.NewSurface(ctx.Max.Width, ctx.Max.Height, pv)

	// Header row
	header := richtext.New([]vaxis.Segment{
		{Text: fmt.Sprintf("%-20s%-10s%10s%10s%10s", "NAME", "STATUS", "SIZE", "ALLOC", "FREE"),
			Style: vaxis.Style{Attribute: vaxis.AttrBold}},
	})
	headerSurf, err := header.Draw(ctx.WithMax(vxfw.Size{Width: ctx.Max.Width, Height: 1}))
	if err != nil {
		return vxfw.Surface{}, err
	}
	s.AddChild(0, 0, headerSurf)

	// List
	listCtx := ctx.WithMax(vxfw.Size{Width: ctx.Max.Width, Height: ctx.Max.Height - 1})
	listSurf, err := pv.list.Draw(listCtx)
	if err != nil {
		return vxfw.Surface{}, err
	}
	s.AddChild(0, 1, listSurf)

	return s, nil
}

// HandleEvent delegates to the list widget for navigation.
func (pv *PoolsView) HandleEvent(ev vaxis.Event, phase vxfw.EventPhase) (vxfw.Command, error) {
	return pv.list.HandleEvent(ev, phase)
}
