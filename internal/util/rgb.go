package util

import "image/color"

func RGB1(rgb int) color.RGBA {
	return color.RGBA{
		R: (uint8)(rgb >> 16),
		G: (uint8)((rgb >> 8) & 0xff),
		B: (uint8)((rgb) & 0xff),
		A: 255,
	}
}
