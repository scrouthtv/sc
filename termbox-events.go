// termbox-events
package main

import (
	"bytes"
	"fmt"
	"os"
	"time"

	"github.com/nsf/termbox-go"
)

type SheetMode int

const (
	NORMAL_MODE SheetMode = iota
	INSERT_MODE SheetMode = iota
	EXIT_MODE   SheetMode = iota
)

func processTermboxEvents(s *Sheet) {
	prompt := ""
	stringEntry := false
	smode := NORMAL_MODE
	valBuffer := bytes.Buffer{}
	insAlign := AlignRight

	// Display
	go func() {
		for _ = range time.Tick(200 * time.Millisecond) {
			switch smode {
			case NORMAL_MODE:
				selSel, _ := s.getCell(s.selectedCell)
				displayValue(fmt.Sprintf("%s (10 2 0) [%s]", s.selectedCell, selSel.statusBarVal()), 0, 0, 80, AlignLeft, false)
			case INSERT_MODE:
				displayValue(fmt.Sprintf("i> %s %s = %s", prompt, s.selectedCell, valBuffer.String()), 0, 0, 80, AlignLeft, false)
			case EXIT_MODE:
				displayValue(fmt.Sprintf("File \"%s\" is modified, save before exiting?", s.filename), 0, 0, 80, AlignLeft, false)
			}
			termbox.Flush()
		}
	}()

	// Events
	for ev := termbox.PollEvent(); ev.Type != termbox.EventError; ev = termbox.PollEvent() {
		switch ev.Type {
		case termbox.EventKey:
			switch smode {
			case NORMAL_MODE:
				switch ev.Key {
				case termbox.KeyArrowUp:
					s.MoveUp()
				case termbox.KeyArrowDown:
					s.MoveDown()
				case termbox.KeyArrowLeft:
					s.MoveLeft()
				case termbox.KeyArrowRight:
					s.MoveRight()
				case 0:
					switch ev.Ch {
					case 'q':
						smode = EXIT_MODE
					case '=', 'i':
						smode = INSERT_MODE
						prompt = "let"
						insAlign = AlignRight
					case '<':
						prompt = "leftstring"
						smode = INSERT_MODE
						insAlign = AlignLeft
						stringEntry = true
					case '>':
						prompt = "rightstring"
						smode = INSERT_MODE
						insAlign = AlignRight
						stringEntry = true
					case '\\':
						prompt = "label"
						smode = INSERT_MODE
						insAlign = AlignCenter
						stringEntry = true
					case 'h':
						s.MoveLeft()
					case 'j':
						s.MoveDown()
					case 'k':
						s.MoveUp()
					case 'l':
						s.MoveRight()
					case 'x':
						s.clearCell(s.selectedCell)
					}
				}
			case INSERT_MODE:
				if ev.Key == termbox.KeyEnter {
					s.setCell(s.selectedCell, &Cell{value: valBuffer.String(), alignment: insAlign, stringType: stringEntry})
					valBuffer.Reset()
					smode = NORMAL_MODE
					stringEntry = false
				} else if ev.Key == termbox.KeyEsc {
					valBuffer.Reset()
					smode = NORMAL_MODE
					stringEntry = false
				} else if ev.Key == termbox.KeyBackspace {
					s := valBuffer.String()
					valBuffer = bytes.Buffer{}
					if len(s) > 0 {
						s = s[0 : len(s)-1]
					}
					valBuffer.WriteString(s)
				} else {
					valBuffer.WriteRune(ev.Ch)
				}
			case EXIT_MODE:
				if ev.Key == 0 && ev.Ch == 'y' {
					if outfile, err := os.Create(s.filename); err == nil {
						fmt.Fprintln(outfile, "# This data file was generated by Spreadsheet Calculator IMproved.")
						fmt.Fprintln(outfile, "# You almost certainly shouldn't edit it.")
						fmt.Fprintln(outfile, "")

						for addr, cell := range s.data {
							cell.write(outfile, addr)
						}
						fmt.Fprintf(outfile, "goto %s A0", s.selectedCell)
						outfile.Close()
					}
				}
				termbox.Close()
				return
			}
		}
	}
}
