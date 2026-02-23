package views_test

import (
	"git.sr.ht/~rockorager/vaxis"
	"git.sr.ht/~rockorager/vaxis/vxfw"
)

func testDrawContext(w, h uint16) vxfw.DrawContext {
	return vxfw.DrawContext{
		Max: vxfw.Size{Width: w, Height: h},
		Min: vxfw.Size{},
		Characters: func(s string) []vaxis.Character {
			chars := make([]vaxis.Character, 0, len(s))
			for _, r := range s {
				chars = append(chars, vaxis.Character{Grapheme: string(r), Width: 1})
			}
			return chars
		},
	}
}
