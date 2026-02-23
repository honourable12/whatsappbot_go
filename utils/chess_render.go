package utils

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"strings"
)

func RenderChessBoard(fen string) ([]byte, error) {
	const (
		sqSize     = 60
		boardSize  = sqSize * 8
		padding    = 30
		totalSize  = boardSize + padding
	)

	// Image with padding for labels (Bottom and Left)
	img := image.NewRGBA(image.Rect(0, 0, totalSize, totalSize))
	bg := color.RGBA{40, 40, 40, 255}
	fillrect_chess(img, 0, 0, totalSize, totalSize, bg)

	// Colors
	light := color.RGBA{240, 217, 181, 255}
	dark := color.RGBA{181, 136, 99, 255}
	whitePiece := color.RGBA{255, 255, 255, 255}
	blackPiece := color.RGBA{0, 0, 0, 255}
	labelColor := color.RGBA{200, 200, 200, 255}

	// Draw squares (shifted by padding/2 for top/right if needed, but let's keep it simple: labels on left and bottom)
	// Board starts at (padding, 0) for numbers on left, and ends at (totalSize, boardSize) for letters on bottom.
	const boardX = padding
	const boardY = 0

	for y := 0; y < 8; y++ {
		// Draw rank labels (1-8) on the left
		drawChar(img, 10, y*sqSize + sqSize/2 - 5, fmt.Sprintf("%d", 8-y), labelColor)

		for x := 0; x < 8; x++ {
			c := light
			if (x+y)%2 == 1 {
				c = dark
			}
			fillrect_chess(img, boardX+x*sqSize, boardY+y*sqSize, boardX+(x+1)*sqSize, boardY+(y+1)*sqSize, c)
		}
	}

	// Draw file labels (a-h) on the bottom
	for x := 0; x < 8; x++ {
		drawChar(img, boardX+x*sqSize + sqSize/2 - 5, boardSize + 10, string('a'+x), labelColor)
	}

	// Parse FEN for pieces
	parts := strings.Split(fen, " ")
	rows := strings.Split(parts[0], "/")

	for r, row := range rows {
		c := 0
		for _, char := range row {
			if char >= '1' && char <= '8' {
				c += int(char - '0')
			} else {
				pieceCol := blackPiece
				if char >= 'A' && char <= 'Z' {
					pieceCol = whitePiece
				}
				
				drawPiece(img, boardX+c*sqSize, boardY+r*sqSize, sqSize, string(char), pieceCol)
				c++
			}
		}
	}

	var buf bytes.Buffer
	err := png.Encode(&buf, img)
	return buf.Bytes(), err
}

func drawChar(img *image.RGBA, x, y int, char string, col color.RGBA) {
	// Simple 5x7 bitmap representation for coordinates
	// Numbers 1-8 and letters a-h
	bitmaps := map[string][]string{
		"1": {"..#..", ".##..", "..#..", "..#..", "..#..", "..#..", ".###."},
		"2": {".###.", "#...#", "....#", "..##.", ".#...", "#....", "#####"},
		"3": {"#####", "....#", "...#.", "..##.", "....#", "#...#", ".###."},
		"4": {"#...#", "#...#", "#...#", "#####", "....#", "....#", "....#"},
		"5": {"#####", "#....", "####.", "....#", "....#", "#...#", ".###."},
		"6": {".###.", "#....", "####.", "#...#", "#...#", "#...#", ".###."},
		"7": {"#####", "....#", "...#.", "..#..", ".#...", ".#...", ".#..."},
		"8": {".###.", "#...#", "#...#", ".###.", "#...#", "#...#", ".###."},
		"a": {".###.", "....#", ".####", "#...#", "#...#", "#...#", ".####"},
		"b": {"#....", "#....", "####.", "#...#", "#...#", "#...#", "####."},
		"c": {".###.", "#...#", "#....", "#....", "#....", "#...#", ".###."},
		"d": {"....#", "....#", ".####", "#...#", "#...#", "#...#", ".####"},
		"e": {".###.", "#...#", "#####", "#....", "#....", "#...#", ".###."},
		"f": {".###.", "#....", "###..", "#....", "#....", "#....", "#...."},
		"g": {".###.", "#...#", "#....", "####.", "....#", "#...#", ".###."},
		"h": {"#....", "#....", "####.", "#...#", "#...#", "#...#", "#...#"},
	}

	if bitmap, ok := bitmaps[char]; ok {
		for r, row := range bitmap {
			for c, pixel := range row {
				if pixel == '#' {
					img.Set(x+c, y+r, col)
					// Make it slightly thicker for visibility
					img.Set(x+c+1, y+r, col)
				}
			}
		}
	}
}

func drawPiece(img *image.RGBA, x, y, size int, name string, col color.RGBA) {
	// Simple circle with letter for now as I don't have piece assets
	centerX, centerY := x+size/2, y+size/2
	radius := size/3
	
	// Inner circle
	for dx := -radius; dx <= radius; dx++ {
		for dy := -radius; dy <= radius; dy++ {
			if dx*dx+dy*dy <= radius*radius {
				img.Set(centerX+dx, centerY+dy, col)
			}
		}
	}

	// Add a little dot in center to differentiate pieces?
	// This is hard without a font.
}

func fillrect_chess(img *image.RGBA, x1, y1, x2, y2 int, c color.Color) {
	for x := x1; x < x2; x++ {
		for y := y1; y < y2; y++ {
			img.Set(x, y, c)
		}
	}
}
