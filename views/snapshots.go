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

// SnapshotsView displays a list of TrueNAS ZFS snapshots.
type SnapshotsView struct {
	service   truenas.SnapshotServiceAPI
	snapshots []truenas.Snapshot
	list      list.Dynamic
}

// NewSnapshotsView creates a SnapshotsView backed by the given service.
func NewSnapshotsView(svc truenas.SnapshotServiceAPI) *SnapshotsView {
	sv := &SnapshotsView{service: svc}
	sv.list.DrawCursor = true
	sv.list.Builder = sv.buildItem
	return sv
}

// Load fetches snapshots from the service.
func (sv *SnapshotsView) Load(ctx context.Context) error {
	snapshots, err := sv.service.List(ctx)
	if err != nil {
		return err
	}
	sv.snapshots = snapshots
	return nil
}

// Snapshots returns the currently loaded snapshots.
func (sv *SnapshotsView) Snapshots() []truenas.Snapshot {
	return sv.snapshots
}

// ItemCount returns the number of loaded snapshots.
func (sv *SnapshotsView) ItemCount() int {
	return len(sv.snapshots)
}

// SelectedSnapshot returns the currently selected snapshot, or nil if empty.
func (sv *SnapshotsView) SelectedSnapshot() *truenas.Snapshot {
	idx := int(sv.list.Cursor())
	if idx >= len(sv.snapshots) {
		return nil
	}
	return &sv.snapshots[idx]
}

func (sv *SnapshotsView) buildItem(i uint, cursor uint) vxfw.Widget {
	if int(i) >= len(sv.snapshots) {
		return nil
	}
	snap := sv.snapshots[i]

	hold := " "
	if snap.HasHold {
		hold = "H"
	}

	return richtext.New([]vaxis.Segment{
		{Text: fmt.Sprintf("%s ", hold)},
		{Text: fmt.Sprintf("%-30s", snap.Dataset)},
		{Text: fmt.Sprintf("%-25s", snap.SnapshotName)},
		{Text: fmt.Sprintf("%10s", humanize.IBytes(uint64(snap.Used)))},
		{Text: fmt.Sprintf("%10s", humanize.IBytes(uint64(snap.Referenced)))},
	})
}

// Draw renders the snapshots list.
func (sv *SnapshotsView) Draw(ctx vxfw.DrawContext) (vxfw.Surface, error) {
	s := vxfw.NewSurface(ctx.Max.Width, ctx.Max.Height, sv)

	header := richtext.New([]vaxis.Segment{
		{Text: fmt.Sprintf("  %-30s%-25s%10s%10s", "DATASET", "SNAPSHOT", "USED", "REFER"),
			Style: vaxis.Style{Attribute: vaxis.AttrBold}},
	})
	headerSurf, err := header.Draw(ctx.WithMax(vxfw.Size{Width: ctx.Max.Width, Height: 1}))
	if err != nil {
		return vxfw.Surface{}, err
	}
	s.AddChild(0, 0, headerSurf)

	listCtx := ctx.WithMax(vxfw.Size{Width: ctx.Max.Width, Height: ctx.Max.Height - 1})
	listSurf, err := sv.list.Draw(listCtx)
	if err != nil {
		return vxfw.Surface{}, err
	}
	s.AddChild(0, 1, listSurf)

	return s, nil
}

// HandleEvent delegates to the list widget for navigation.
func (sv *SnapshotsView) HandleEvent(ev vaxis.Event, phase vxfw.EventPhase) (vxfw.Command, error) {
	return sv.list.HandleEvent(ev, phase)
}
