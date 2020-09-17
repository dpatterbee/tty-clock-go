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
	xOffset       int
	yOffset       int
	terminalSizeX int
	terminalSizeY int
	displaySizeX  int
	displaySizeY  int
	dateOffset    int
	sync.RWMutex
}

type coord struct {
	x int
	y int
}

var defStyle, onStyle tcell.Style

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

	defStyle = tcell.StyleDefault.Background(tcell.ColorBlack).Foreground(tcell.ColorRed)
	s.SetStyle(defStyle)
	onStyle = tcell.StyleDefault.Background(tcell.ColorRed).Foreground(tcell.ColorBlack)
	options.Lock()
	flags.Parse(&options)
	options.terminalSizeX, options.terminalSizeY = s.Size()
	options.yOffset = 1
	options.Unlock()

	forceUpdate := make(chan bool)

	go func() {
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
	}()

	drawClock(s, forceUpdate)
}

func setCenter() {
	options.Lock()
	defer options.Unlock()

	if !options.Center {
		return
	}

	xPos := (options.terminalSizeX - options.displaySizeX) / 2
	yPos := (options.terminalSizeY - options.displaySizeY) / 2

	options.xOffset = xPos - 1
	options.yOffset = yPos + 2
}

func drawClock(s tcell.Screen, forceUpdateChan chan bool) {
	var timeWait time.Time
	for {
		currTime := time.Now()
		timeWait = currTime.Add(time.Second / 2).Round(time.Second)

		options.RLock()
		clockTime := currTime.Format(timeFormats[options.TwelveHour][options.Seconds])
		clockDate := currTime.Format(dateFormats[options.TwelveHour])
		options.RUnlock()
		displayMatrix := parseArea(clockTime)

		// Center the clock if necessary
		setCenter()

		drawArea(s, displayMatrix, clockDate)
		s.Show()

		select {
		case <-forceUpdateChan:
		case <-time.After(time.Until(timeWait)):
		}
	}
}

func drawArea(s tcell.Screen, displayMatrix [][]bool, date string) {
	s.Clear()
	options.RLock()
	defer options.RUnlock()
	for j, v := range displayMatrix {
		for i, x := range v {
			if x {
				s.SetContent(options.xOffset+i, options.yOffset+j, ' ', nil, onStyle)
			} else {
				s.SetContent(options.xOffset+i, options.yOffset+j, ' ', nil, defStyle)
			}
		}
	}
	for i, v := range date {
		s.SetContent(options.xOffset+options.displaySizeX/2-5+i, options.yOffset+6, v, nil, defStyle)
	}
}

func parseArea(time string) [][]bool {
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
