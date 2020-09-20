package main

import (
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/gdamore/tcell"
	"github.com/jessevdk/go-flags"
)

var character = map[rune][][]bool{
	'0': {{true, true, true, true, true, true}, {true, true, false, false, true, true}, {true, true, false, false, true, true}, {true, true, false, false, true, true}, {true, true, true, true, true, true}},               /* 0 */
	'1': {{false, false, false, false, true, true}, {false, false, false, false, true, true}, {false, false, false, false, true, true}, {false, false, false, false, true, true}, {false, false, false, false, true, true}}, /* 1 */
	'2': {{true, true, true, true, true, true}, {false, false, false, false, true, true}, {true, true, true, true, true, true}, {true, true, false, false, false, false}, {true, true, true, true, true, true}},             /* 2 */
	'3': {{true, true, true, true, true, true}, {false, false, false, false, true, true}, {true, true, true, true, true, true}, {false, false, false, false, true, true}, {true, true, true, true, true, true}},             /* 3 */
	'4': {{true, true, false, false, true, true}, {true, true, false, false, true, true}, {true, true, true, true, true, true}, {false, false, false, false, true, true}, {false, false, false, false, true, true}},         /* 4 */
	'5': {{true, true, true, true, true, true}, {true, true, false, false, false, false}, {true, true, true, true, true, true}, {false, false, false, false, true, true}, {true, true, true, true, true, true}},             /* 5 */
	'6': {{true, true, true, true, true, true}, {true, true, false, false, false, false}, {true, true, true, true, true, true}, {true, true, false, false, true, true}, {true, true, true, true, true, true}},               /* 6 */
	'7': {{true, true, true, true, true, true}, {false, false, false, false, true, true}, {false, false, false, false, true, true}, {false, false, false, false, true, true}, {false, false, false, false, true, true}},     /* 7 */
	'8': {{true, true, true, true, true, true}, {true, true, false, false, true, true}, {true, true, true, true, true, true}, {true, true, false, false, true, true}, {true, true, true, true, true, true}},                 /* 8 */
	'9': {{true, true, true, true, true, true}, {true, true, false, false, true, true}, {true, true, true, true, true, true}, {false, false, false, false, true, true}, {true, true, true, true, true, true}},               /* 9 */
	':': {{false, false, false, false}, {false, true, true, false}, {false, false, false, false}, {false, true, true, false}, {false, false, false, false}},                                                                 /* : */
}

var timeFormats = map[bool]map[bool]string{
	true: {true: "03:04:05", false: "03:04"}, false: {true: "15:04:05", false: "15:04"},
}
var dateFormats = map[bool]string{
	true: "2006-01-02 [PM]", false: "2006-01-02",
}

var options struct {
	Seconds       bool `short:"s" description:"Display Seconds"`
	Center        bool `short:"c" description:"Center the clock"`
	TwelveHour    bool `short:"t" description:"Display in 12 hour format"`
	Colour        int  `short:"C" default:"2" description:"Choose clock colour [1-378]"`
	xOffset       int
	yOffset       int
	terminalSizeX int
	terminalSizeY int
	displaySizeX  int
	displaySizeY  int
	dateOffset    int
	defStyle      tcell.Style
	onStyle       tcell.Style
	sync.RWMutex
}

type coord struct {
	x int
	y int
}

func main() {

	s, e := tcell.NewScreen()
	if e != nil {
		fmt.Fprintf(os.Stderr, "%v\n", e)
		os.Exit(1)
	}
	if e := s.Init(); e != nil {
		fmt.Fprintf(os.Stderr, "%v\n", e)
		os.Exit(1)
	}

	options.Lock()
	flags.Parse(&options)

	if options.Colour > 378 || options.Colour < 1 {
		options.Colour = 2
	}
	options.defStyle = tcell.StyleDefault.Foreground(tcell.Color(options.Colour))
	options.onStyle = tcell.StyleDefault.Background(tcell.Color(options.Colour))
	s.SetStyle(options.defStyle)

	options.terminalSizeX, options.terminalSizeY = s.Size()
	options.yOffset = 1
	options.Unlock()

	forceUpdate := make(chan bool)

	go handleInput(s, forceUpdate)

	// Draws initial clock before main loop setup so that it appears as soon as the program launches and not up to 1 second later.
	drawClock(s, time.Now())

	// Main program loop lives here
	updateClock(s, forceUpdate)
}

// handleInput asynchronously polls for input events and handles each accordingly, forcing an update to the clock face when required.
func handleInput(s tcell.Screen, forceUpdate chan bool) {
	for {
		ev := s.PollEvent()
		// You like switches? We got em
		switch ev := ev.(type) {

		case *tcell.EventResize:
			options.Lock()
			options.terminalSizeX, options.terminalSizeY = s.Size()
			options.Unlock()
			forceUpdate <- true

		case *tcell.EventKey:
			switch ev.Key() {

			case tcell.KeyEscape:
				s.Fini()
				os.Exit(0)

			case tcell.KeyDown:
				options.Lock()
				options.yOffset++
				options.Unlock()
				forceUpdate <- true

			case tcell.KeyRune:
				switch ev.Rune() {

				case 'q':
					s.Fini()
					os.Exit(0)

				case 't', 'T':
					options.Lock()
					options.TwelveHour = !options.TwelveHour
					options.Unlock()
					forceUpdate <- true
				case 's', 'S':

					options.Lock()
					options.Seconds = !options.Seconds
					options.Unlock()
					forceUpdate <- true
				case 'c', 'C':

					options.Lock()
					options.Center = !options.Center
					options.Unlock()
					forceUpdate <- true

				case 'h':
					options.Lock()
					if !options.Center && options.xOffset > 0 {
						options.xOffset--
					}
					options.Unlock()
					forceUpdate <- true

				case 'j':
					options.Lock()
					if !options.Center && options.yOffset < options.terminalSizeY-options.displaySizeY-1 {
						options.yOffset++
					}
					options.Unlock()
					forceUpdate <- true

				case 'k':
					options.Lock()
					if !options.Center && options.yOffset > 1 {
						options.yOffset--
					}
					options.Unlock()
					forceUpdate <- true

				case 'l':
					options.Lock()
					if !options.Center && options.xOffset < options.terminalSizeX-options.displaySizeX-1 {
						options.xOffset++
					}
					options.Unlock()
					forceUpdate <- true

				}
			}
		}
	}
}

// setCenter checks if the "Center" option is set and, if so, sets the clock X and Y offset to position the clock centrally
func setCenter() {
	options.Lock()
	defer options.Unlock()

	if !options.Center {
		return
	}

	options.xOffset = (options.terminalSizeX-options.displaySizeX)/2 - 1
	options.yOffset = (options.terminalSizeY-options.displaySizeY)/2 + 2
}

// updateClock starts a loop at the beginning of the second after it is called which calls the clock updating function.
// It loops anytime there is input or when a second has passed.
func updateClock(s tcell.Screen, forceUpdateChan chan bool) {
	time.Sleep(time.Until(time.Now().Add(time.Second / 2).Round(time.Second)))

	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	currTime := time.Now()

	for {
		drawClock(s, currTime)

		// Waits for next tick of ticker or next forced update,
		select {
		case currTime = <-ticker.C:
		case <-forceUpdateChan:
		}
	}
}

// drawClock takes the current time, formats and positions it according to the current settings, and displays it on the terminal screen.
func drawClock(s tcell.Screen, t time.Time) {
	options.RLock()
	clockTime := t.Format(timeFormats[options.TwelveHour][options.Seconds])
	clockDate := t.Format(dateFormats[options.TwelveHour])
	options.RUnlock()

	displayMatrix := clockMatrix(clockTime)

	setCenter()
	s.Clear()
	options.RLock()
	for i, v := range displayMatrix {
		for j, w := range v {
			if w {
				s.SetContent(options.xOffset+j, options.yOffset+i, ' ', nil, options.onStyle)
			} else {
				s.SetContent(options.xOffset+j, options.yOffset+i, ' ', nil, options.defStyle)
			}
		}
	}
	for i, v := range clockDate {
		s.SetContent(options.xOffset+options.displaySizeX/2-5+i, options.yOffset+6, v, nil, options.defStyle)
	}
	options.RUnlock()
	s.Show()
}

// clockMatrix returns a boolean matrix which represents the clock face where true represents "on" cells and false represents "off" cells.
func clockMatrix(time string) [][]bool {
	output := make([][]bool, 5)

	for _, v := range time {
		char := character[v]
		for i := range output {
			output[i] = append(output[i], false)
		}
		for i, x := range char {
			output[i] = append(output[i], x...)
		}

	}

	length := 0
	for _, v := range output {
		if len(v) > length {
			length = len(v)
		}
	}

	options.Lock()
	options.displaySizeX = length + 1
	options.displaySizeY = len(output) + 2

	if options.xOffset+options.displaySizeX > options.terminalSizeX {
		if options.displaySizeX > options.terminalSizeX {
			options.xOffset = 0
		} else {
			options.xOffset = options.terminalSizeX - options.displaySizeX
		}
	}
	if options.yOffset+options.displaySizeY > options.terminalSizeY {
		if options.displaySizeY > options.terminalSizeY {
			options.yOffset = 1
		} else {
			options.yOffset = options.terminalSizeY - options.displaySizeY
		}
	}
	options.Unlock()

	return output

}
