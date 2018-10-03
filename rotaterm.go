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
	Matrix [][]rune
	SizeX  int
	SizeY  int
}

func newTerm(x, y int) *Term {
	data := make([][]rune, x+1)
	for i := range data {
		data[i] = make([]rune, y+1)
	}
	return &Term{data, x, y}
}

// draws time stats
func (screen *Term) stats(s tcell.Screen, e1 time.Duration, e2 time.Duration, e3 time.Duration, sY int) {
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
		s.SetCell(i+1, sY-3, tcell.StyleDefault.Foreground(tcell.Color111).Background(mid1), rune(line1[i]))
	}
	for i := 0; i < len(line2); i++ {
		s.SetCell(i+1, sY-2, tcell.StyleDefault.Foreground(tcell.ColorGreen).Background(mid2), rune(line2[i]))
	}
	for i := 0; i < len(line3); i++ {
		s.SetCell(i+1, sY-1, tcell.StyleDefault.Foreground(tcell.ColorYellow).Background(mid3), rune(line3[i]))
	}
}

// draw key modifier  stats, mostly copy/paste of above ¯\_(ツ)_/¯
func (screen *Term) modstats(s tcell.Screen, radius int, circles int, offset int, sY int) {
	line1 := fmt.Sprintf("radius [A-Z]: %v", radius)
	line2 := fmt.Sprintf("circles[S-X]: %v", circles)
	line3 := fmt.Sprintf("offset [D-C]: %v", offset)

	mid1 := tcell.NewRGBColor(43, 43, 43)
	mid2 := tcell.NewRGBColor(40, 40, 40)
	mid3 := tcell.NewRGBColor(30, 30, 30)

	for i := 0; i < len(line1); i++ {
		s.SetCell(i+20, sY-3, tcell.StyleDefault.Foreground(tcell.Color111).Background(mid1), rune(line1[i]))
	}
	for i := 0; i < len(line2); i++ {
		s.SetCell(i+20, sY-2, tcell.StyleDefault.Foreground(tcell.ColorGreen).Background(mid2), rune(line2[i]))
	}
	for i := 0; i < len(line3); i++ {
		s.SetCell(i+20, sY-1, tcell.StyleDefault.Foreground(tcell.ColorYellow).Background(mid3), rune(line3[i]))
	}
}

// reads screen.Matrix for display
func (screen *Term) draw(s tcell.Screen) {
	st := tcell.StyleDefault
	for x, _ := range screen.Matrix {
		for y, _ := range screen.Matrix[x] {
			s.SetCell(x, y, st, screen.Matrix[x][y])
		}
	}
}

// writes screen.Matrix with a braille image cache
func (screen *Term) makeScreen(m image.Image, s tcell.Screen) {
	if screen.SizeX == 0 || screen.SizeY == 0 {
		return
	}
	buf := new(bytes.Buffer)
	dotmatrix.Print(buf, m)

	for x, _ := range screen.Matrix {
		for y, _ := range screen.Matrix[x] {
			screen.Matrix[x][y] = ' '
		}
	}

	for y := 0; y <= screen.SizeY; y++ {
		for x := 0; x <= screen.SizeX; x++ {
			if buf.Len() > 0 {
				gl, _, e := buf.ReadRune()
				if e != nil {
					fmt.Fprintf(os.Stderr, "%#v\n", e)
				}
				if gl == '\n' {
					screen.Matrix[x][y] = ' '
					break
				}
				screen.Matrix[x][y] = gl
			}
		}
	}
}

func main() {
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

	sX, sY := s.Size()
	screen := newTerm(s.Size())
	dc := gg.NewContext(sX*2, sY*4)

	radius := 10.00
	circles := 400
	offset := 0
	moveX := 0
	moveY := 0

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
					radius = 10.0
					circles = 500.0
					offset = 0
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
						radius--
					}
					if ev.Rune() == 's' || ev.Rune() == 'S' {
						circles = circles + 2
					}
					if ev.Rune() == 'x' || ev.Rune() == 'X' {
						circles = circles - 2
					}
					if ev.Rune() == 'd' || ev.Rune() == 'D' {
						offset = offset + 2
					}
					if ev.Rune() == 'c' || ev.Rune() == 'C' {
						offset = offset - 2
					}
				}
			case *tcell.EventResize:
				sX, sY = s.Size()
				screen = newTerm(s.Size())
				dc = gg.NewContext(sX*2, sY*4)
				s.Sync()
			}
		}
	}()

	rotate := -1000.0
	elapsed3 := time.Duration(0)
	start := time.Now()
	t := time.Now()

loop:
	for {
		select {
		case <-quit:
			break loop
			// minimum loop speed ~30fps
		case <-time.After(time.Millisecond * 33):
		}
		// record previous lap, start new image
		t = time.Now()
		elapsed3 = t.Sub(start) - (time.Millisecond * 33)
		start = time.Now()

		// create an image with lots of circles
		dc.SetRGB(1, 1, 1)
		dc.Clear()
		dc.SetRGB(0, 0, 0)
		for i := 0; i <= circles; i++ {
			t := float64(i) / (400 + rotate/9)
			d := t*rotate*0.6 + 10 + float64(offset)
			a := t * math.Pi * 2 * 20
			x := float64(screen.SizeX+moveX) + math.Cos(a)*d
			y := float64(screen.SizeX-screen.SizeY/2+moveY) + math.Sin(a)*d
			r := t * radius
			dc.DrawCircle(x, y, r)
		}
		dc.Fill()
		ms := dc.Image()

		t = time.Now()
		elapsed1 := t.Sub(start)

		// convert image to pixel Matrix data for display
		start = time.Now()
		screen.makeScreen(ms, s)
		screen.draw(s)
		t = time.Now()
		elapsed2 := t.Sub(start)

		// elapsed3 is the measured time of the previous frame
		screen.stats(s, elapsed1, elapsed2, elapsed3, sY)
		screen.modstats(s, int(radius), circles, offset, sY)

		// update the screen
		s.Show()

		// increment rotate for next frame
		rotate = rotate + 2
		if rotate > 800 {
			rotate = -1000
		}
	}
	s.Fini()
}
