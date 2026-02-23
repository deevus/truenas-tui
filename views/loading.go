package views

import (
	"git.sr.ht/~rockorager/vaxis"
	"git.sr.ht/~rockorager/vaxis/vxfw"
	"git.sr.ht/~rockorager/vaxis/vxfw/richtext"
)

// drawLoadingState renders a "Loading..." message in the view.
func drawLoadingState(ctx vxfw.DrawContext, owner vxfw.Widget) (vxfw.Surface, error) {
	s := vxfw.NewSurface(ctx.Max.Width, ctx.Max.Height, owner)
	label := richtext.New([]vaxis.Segment{
		{Text: "Loading...", Style: vaxis.Style{Attribute: vaxis.AttrDim}},
	})
	labelSurf, err := label.Draw(ctx.WithMax(vxfw.Size{Width: ctx.Max.Width, Height: 1}))
	if err != nil {
		return vxfw.Surface{}, err
	}
	s.AddChild(0, 0, labelSurf)
	return s, nil
}
