// Create an image object and render it to a terminal window using braille characters
// Try to be fast and not care too much about using CPU. Can be used as a test for terminal
// display speed.

// Everything in one file for convenience, this is just a base for further experimentation.
// The basic idea is to use image generation algos to feed to a braille printer for animation.

// Key modifiers are displayed on-screen, additionally:
// CTRL-L: restarts with sane defaults
// ESC/Enter: exits
// Arrow keys: move X/Y coordinates
// https://github.com/onnos/rotaterm

package main

import (
	"bytes"
	"fmt"
	"image"
	"math"
	"os"
	"time"

	"github.com/fogleman/gg"
	"github.com/gdamore/tcell"
	"github.com/kevin-cantwell/dotmatrix"
)

// a terminal window matrix with Runes
type Term struct {
	matrix [][]rune
	sizeX  int
	sizeY  int
}

func newTerm(x, y int) *Term {
	data := make([][]rune, x+1)
	for i := range data {
		data[i] = make([]rune, y+1)
	}
	return &Term{data, x, y}
}

// draws time stats
func (screen *Term) stats(s tcell.Screen, e1 time.Duration, e2 time.Duration, e3 time.Duration) {
	line1 := fmt.Sprintf("mkimg:  %3v", e1.Truncate(time.Millisecond))
	line2 := fmt.Sprintf("mkdots: %3v", e2.Truncate(time.Millisecond))
	line3 := fmt.Sprintf("screen: %3v", e3.Truncate(time.Millisecond))

	mid1 := tcell.NewRGBColor(43, 43, 43)
	mid2 := tcell.NewRGBColor(40, 40, 40)
	mid3 := tcell.NewRGBColor(30, 30, 30)
	if e3 > (time.Millisecond * 30) {
		mid3 = tcell.NewRGBColor(255, 0, 0)
	}

	for i := 0; i < len(line1); i++ {
		s.SetContent(i+1, screen.sizeY-3, rune(line1[i]), nil, tcell.StyleDefault.Foreground(tcell.Color111).Background(mid1))
	}
	for i := 0; i < len(line2); i++ {
		s.SetContent(i+1, screen.sizeY-2, rune(line2[i]), nil, tcell.StyleDefault.Foreground(tcell.ColorGreen).Background(mid2))
	}
	for i := 0; i < len(line3); i++ {
		s.SetContent(i+1, screen.sizeY-1, rune(line3[i]), nil, tcell.StyleDefault.Foreground(tcell.ColorYellow).Background(mid3))
	}
}

// draw key modifier  stats, mostly copy/paste of above ¯\_(ツ)_/¯
func (screen *Term) modstats(s tcell.Screen, radius int, circles int, offset int) {
	line1 := fmt.Sprintf("radius [A-Z]: %v", radius)
	line2 := fmt.Sprintf("circles[S-X]: %v", circles)
	line3 := fmt.Sprintf("offset [D-C]: %v", offset)

	mid1 := tcell.NewRGBColor(43, 43, 43)
	mid2 := tcell.NewRGBColor(40, 40, 40)
	mid3 := tcell.NewRGBColor(30, 30, 30)

	for i := 0; i < len(line1); i++ {
		s.SetContent(i+20, screen.sizeY-3, rune(line1[i]), nil, tcell.StyleDefault.Foreground(tcell.Color111).Background(mid1))
	}
	for i := 0; i < len(line2); i++ {
		s.SetContent(i+20, screen.sizeY-2, rune(line2[i]), nil, tcell.StyleDefault.Foreground(tcell.ColorGreen).Background(mid2))
	}
	for i := 0; i < len(line3); i++ {
		s.SetContent(i+20, screen.sizeY-1, rune(line3[i]), nil, tcell.StyleDefault.Foreground(tcell.ColorYellow).Background(mid3))
	}
}

// reads screen.matrix for display
func (screen *Term) draw(s tcell.Screen) {
	st := tcell.StyleDefault
	for x, _ := range screen.matrix {
		for y, _ := range screen.matrix[x] {
			s.SetContent(x, y, screen.matrix[x][y], nil, st)
		}
	}
}

// writes screen.matrix with a braille image cache
func (screen *Term) makeScreen(m image.Image, s tcell.Screen) {
	if screen.sizeX == 0 || screen.sizeY == 0 {
		return
	}
	buf := new(bytes.Buffer)
	dotmatrix.Print(buf, m)

	for y := 0; y <= screen.sizeY; y++ {
		for x := 0; x <= screen.sizeX; x++ {
			if buf.Len() > 0 {
				gl, _, e := buf.ReadRune()
				if e != nil {
					fmt.Fprintf(os.Stderr, "%#v\n", e)
				}
				if gl == '\n' {
					break
				}
				screen.matrix[x][y] = gl
			}
		}
	}
	screen.draw(s)
}
func (screen *Term) makeImage(dc gg.Context, rotate float64, radius float64, circles int, offset int, moveX int, moveY int) {
	dc.SetRGB(1, 1, 1)
	dc.Clear()
	dc.SetRGB(0, 0, 0)
	for i := 0; i <= circles; i++ {
		t := float64(i) / (400 + rotate/9)
		d := t*rotate*0.6 + 10 + float64(offset)
		a := t * math.Pi * 2 * 20
		x := float64(screen.sizeX+moveX) + math.Cos(a)*d
		y := float64(screen.sizeX-screen.sizeY/2+moveY) + math.Sin(a)*d
		r := t * radius
		dc.DrawCircle(x, y, r)
	}
	dc.Fill()
}

func main() {
	// minimum loop speed (1000ms/30fps =~33ms)
	const delay = 33
	var rotate float64
	var elapsed3 time.Duration
	start := time.Now()
	t := time.Now()
	radius := 6.00
	circles := 400
	offset := 10
	moveX := 0
	moveY := 0

	tcell.SetEncodingFallback(tcell.EncodingFallbackASCII)
	s, e := tcell.NewScreen()
	if e != nil {
		fmt.Fprintf(os.Stderr, "%v\n", e)
		os.Exit(1)
	}
	if e = s.Init(); e != nil {
		fmt.Fprintf(os.Stderr, "%v\n", e)
		os.Exit(1)
	}

	s.SetStyle(tcell.StyleDefault.
		Foreground(tcell.Color111).
		Background(tcell.ColorBlack))
	s.Clear()

	screen := newTerm(s.Size())
	dc := gg.NewContext(screen.sizeX*2, screen.sizeY*4)

	quit := make(chan struct{})
	go func() {
		for {
			ev := s.PollEvent()
			switch ev := ev.(type) {
			case *tcell.EventKey:
				switch ev.Key() {
				case tcell.KeyEscape, tcell.KeyEnter:
					close(quit)
					return
				case tcell.KeyCtrlL:
					radius = 6.0
					circles = 400
					offset = 10
					rotate = 0
					moveX = 0
					moveY = 0

					s.Clear()
					s.Sync()
				case tcell.KeyRight:
					moveX++
				case tcell.KeyLeft:
					moveX--
				case tcell.KeyUp:
					moveY--
				case tcell.KeyDown:
					moveY++
				case tcell.KeyRune:
					if ev.Rune() == 'a' || ev.Rune() == 'A' {
						radius++
					}
					if ev.Rune() == 'z' || ev.Rune() == 'Z' {
						if radius > 0 {
							radius--
						}
					}
					if ev.Rune() == 's' || ev.Rune() == 'S' {
						circles += 5
					}
					if ev.Rune() == 'x' || ev.Rune() == 'X' {
						circles -= 5
					}
					if ev.Rune() == 'd' || ev.Rune() == 'D' {
						offset -= 1
					}
					if ev.Rune() == 'c' || ev.Rune() == 'C' {
						offset += 1
					}
				}
			case *tcell.EventResize:
				screen = newTerm(s.Size())
				dc = gg.NewContext(screen.sizeX*2, screen.sizeY*4)
				s.Sync()
			}
		}
	}()

loop:
	for {
		select {
		case <-quit:
			break loop
		case <-time.After(time.Millisecond * delay):
		}
		// record previous lap, start new image
		t = time.Now()
		elapsed3 = t.Sub(start) - (time.Millisecond * delay)
		start = time.Now()

		// create an image with lots of circles
		screen.makeImage(*dc, rotate, radius, circles, offset, moveX, moveY)
		t = time.Now()
		elapsed1 := t.Sub(start)

		// convert image to pixel matrix data for display
		start = time.Now()
		screen.makeScreen(dc.Image(), s)
		t = time.Now()
		elapsed2 := t.Sub(start)

		// elapsed3 is the measured time of the previous frame
		screen.stats(s, elapsed1, elapsed2, elapsed3)
		screen.modstats(s, int(radius), circles, offset)

		// increment rotate for next frame
		rotate = rotate + 2
		if rotate > 800 {
			rotate = -800
		}

		// update the screen
		s.Show()
	}
	s.Fini()
}
