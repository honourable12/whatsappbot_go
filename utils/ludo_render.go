package utils

import (
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"bytes"
)

func GenerateLudoBoard() ([]byte, error) {
	const size = 600
	img := image.NewRGBA(image.Rect(0, 0, size, size))
	
	// Background
	draw.Draw(img, img.Bounds(), &image.Uniform{color.White}, image.Point{}, draw.Src)

	// Draw basic regions
	red := color.RGBA{255, 0, 0, 255}
	blue := color.RGBA{0, 0, 255, 255}
	yellow := color.RGBA{255, 255, 0, 255}
	green := color.RGBA{0, 255, 0, 255}

	// Home bases
	fillrect(img, 0, 0, 240, 240, red)
	fillrect(img, 360, 0, 600, 240, blue)
	fillrect(img, 0, 360, 240, 600, green)
	fillrect(img, 360, 360, 600, 600, yellow)

	// Center
	fillrect(img, 240, 240, 360, 360, color.RGBA{200, 200, 200, 255})

	var buf bytes.Buffer
	err := png.Encode(&buf, img)
	return buf.Bytes(), err
}

func fillrect(img *image.RGBA, x1, y1, x2, y2 int, c color.Color) {
	for x := x1; x < x2; x++ {
		for y := y1; y < y2; y++ {
			img.Set(x, y, c)
		}
	}
}
