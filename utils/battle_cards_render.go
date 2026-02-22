package utils

import (
	"bytes"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"whatsappbot_go/games"
)

func RenderBattleBoard(g *games.BattleGame) ([]byte, error) {
	const (
		width  = 800
		height = 600
		cardW  = 120
		cardH  = 160
	)

	img := image.NewRGBA(image.Rect(0, 0, width, height))
	bg := color.RGBA{30, 30, 30, 255}
	draw.Draw(img, img.Bounds(), &image.Uniform{bg}, image.Point{}, draw.Src)

	// Colors
	red := color.RGBA{200, 50, 50, 255}
	blue := color.RGBA{50, 50, 200, 255}
	green := color.RGBA{50, 200, 50, 255}
	yellow := color.RGBA{200, 200, 50, 255}
	white := color.RGBA{255, 255, 255, 255}

	getColor := func(c string) color.RGBA {
		switch c {
		case "Red": return red
		case "Blue": return blue
		case "Green": return green
		case "Yellow": return yellow
		default: return white
		}
	}

	// Player 1 Area (Top)
	fillrect(img, 10, 10, width-10, 50, color.RGBA{100, 0, 0, 255})
	// Player 2 Area (Bottom)
	fillrect(img, 10, height-50, width-10, height-10, color.RGBA{0, 0, 100, 255})

	// Draw Player 1 Board
	for i, card := range g.Player1.Board {
		x := 50 + i*(cardW+20)
		y := 70
		drawCard(img, x, y, cardW, cardH, card, getColor(card.Color))
	}

	// Draw Player 2 Board
	for i, card := range g.Player2.Board {
		x := 50 + i*(cardW+20)
		y := height - 70 - cardH
		drawCard(img, x, y, cardW, cardH, card, getColor(card.Color))
	}

	// Mid line
	fillrect(img, 0, height/2-1, width, height/2+1, color.RGBA{150, 150, 150, 255})

	var buf bytes.Buffer
	err := png.Encode(&buf, img)
	return buf.Bytes(), err
}

func drawCard(img *image.RGBA, x, y, w, h int, c games.BattleCard, col color.RGBA) {
	// Card Border
	fillrect(img, x, y, x+w, y+h, color.RGBA{255, 255, 255, 255})
	// Card Face
	fillrect(img, x+5, y+5, x+w-5, y+h-5, col)
	
	// Stats bars (Visual representation of ATK/DEF)
	// Attack bar (top)
	atkH := (c.Attack * (h - 20)) / 50
	fillrect(img, x+w-15, y+h-5-atkH, x+w-10, y+h-5, color.RGBA{255, 100, 100, 255})
	
	// Defense bar (side)
	defH := (c.Defense * (h - 20)) / 50
	fillrect(img, x+10, y+h-5-defH, x+15, y+h-5, color.RGBA{100, 100, 255, 255})
}
