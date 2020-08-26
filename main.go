package main

import (
	"fmt"
	"os"
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
	':': {{false, false, false, false}, {false, true, true, false}, {false, false, false, false}, {false, true, true, false}, {false, false, false, false}},
}

var col = [4]int{0, 7, 19, 26}

var options struct {
	Seconds bool `short:"s" description:"Display Seconds"`
	Center  bool `short:"c" description:"Center the clock"`
}
var timeFormat string = "15:04"
var dateFormat string = "2006-01-02"

type coord struct {
	x int
	y int
}

var defStyle, onStyle tcell.Style
var centred bool

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

	flags.Parse(&options)

	if options.Seconds {
		timeFormat = "15:04:05"
	}
	if options.Center {
		centred = true
	}

	sizeChan := make(chan coord)

	go func() {
		for {
			ev := s.PollEvent()
			switch ev := ev.(type) {
			case *tcell.EventResize:
				x, y := s.Size()
				c := coord{x, y}
				sizeChan <- c
			case *tcell.EventKey:
				switch ev.Key() {
				case tcell.KeyEscape:
					s.Fini()
					os.Exit(0)
				case tcell.KeyRune:
					switch ev.Rune() {
					case 'q':
						s.Fini()
						os.Exit(0)
					}
				}
			}
		}
	}()

	drawClock(s, sizeChan)
}

func drawClock(s tcell.Screen, termSizeChan chan coord) {
	x, y := s.Size()
	termSize := coord{x, y}
	timeWait := time.Now().Round(time.Second)
	for {
		clockTime := time.Now().Format(timeFormat)
		clockDate := time.Now().Format(dateFormat)
		displayMatrix := parseArea(clockTime)

		drawArea(s, displayMatrix, termSize, clockDate)
		s.Show()

		select {
		case termSize = <-termSizeChan:
		case <-time.After(time.Until(timeWait)):
			timeWait = time.Now().Round(time.Second).Add(time.Second)
		}
	}
}

func drawArea(s tcell.Screen, displayMatrix [8][]bool, termSize coord, date string) {
	origin, offset := getOriginAndMid(termSize, &displayMatrix)
	s.Clear()
	for j, v := range displayMatrix {
		for i, x := range v {
			if x {
				s.SetContent(origin.x+i, origin.y+j, ' ', nil, onStyle)
			} else {
				s.SetContent(origin.x+i, origin.y+j, ' ', nil, defStyle)
			}
		}
	}
	for i, v := range date {
		s.SetContent(origin.x+offset+i, origin.y+7, v, nil, defStyle)
	}
}

func parseArea(time string) [8][]bool {
	output := [8][]bool{}

	for _, v := range time {
		char := character[v]
		for i := range output {
			output[i] = append(output[i], false)
		}
		for i, x := range char {
			output[i+1] = append(output[i+1], x...)
		}

	}

	return output

}

func getOriginAndMid(termSize coord, displayMatrix *[8][]bool) (coord, int) {
	var center coord
	if centred {
		center = coord{x: (termSize.x-len(displayMatrix[1]))/2 - 1, y: (termSize.y - 7) / 2}
	}
	if center.x < 0 {
		center.x = 0
	}
	if center.y < 0 {
		center.y = 0
	}
	return center, len(displayMatrix[1])/2 - 4
}