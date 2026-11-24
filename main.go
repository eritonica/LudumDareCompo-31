package main

import (
	"errors"
	"fmt"
	"log"

	"github.com/hajimehoshi/ebiten/v2"
)

const (
	screenWidth  = 1920
	screenHeight = 1080
)

func main() {
	fmt.Printf("Theme for the game was given in Ludum Dare 31 and it was voted as Entire Game on One Screen\n")

	snowball, err := NewGame()
	if err != nil {
		log.Fatal(err)
	}

	ebiten.SetWindowTitle("LD31 Snowball")
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)
	ebiten.SetVsyncEnabled(true)
	ebiten.SetTPS(60)

	// The game is laid out for 1920x1080. Fit the window to the monitor;
	// the logical screen is scaled to whatever size the window ends up.
	w, h := screenWidth, screenHeight
	if mw, mh := ebiten.Monitor().Size(); mw > 0 && mh > 0 {
		scale := min(float64(mw)*0.9/screenWidth, float64(mh)*0.9/screenHeight, 1.0)
		w, h = int(screenWidth*scale), int(screenHeight*scale)
	}
	ebiten.SetWindowSize(w, h)

	if err := ebiten.RunGame(snowball); err != nil && !errors.Is(err, ebiten.Termination) {
		log.Fatal(err)
	}
}
