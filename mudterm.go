package main

import (
	"fmt"
	"io"

	"github.com/gdamore/tcell"
	"github.com/gdamore/tcell/terminfo"
)

type cell struct {
	c     rune
	style tcell.Style
}

type terminal struct {
	c             io.ReadWriter
	cols, rows    int
	term          *terminfo.Terminfo
	last, current []cell
	style         tcell.Style
}

func (t *terminal) SetInfo(cols, rows int, term string) error {
	var err error
	t.term, err = terminfo.LookupTerminfo(term)
	if err != nil {
		return err
	}
	t.cols = cols
	t.rows = rows
	t.style = tcell.StyleDefault
	t.last = make([]cell, cols*rows)
	t.current = make([]cell, cols*rows)
	t.Clear()
	return nil
}

func (t *terminal) SetSize(cols, rows int) {
	t.cols = cols
	t.rows = rows
}

func (t *terminal) KeyPress(b byte) {
	fmt.Println("Key Press", string(b))
}

func (t *terminal) ProcessInput() error {
	readBuf := make([]byte, 1)
	n, err := t.c.Read(readBuf)
	if err != nil || n != 1 {
		return err
	}
	t.KeyPress(readBuf[0])
	return nil
}

func newTerminal(c io.ReadWriter) *terminal {
	return &terminal{c: c}
}

// interface for tcell.Screen

func (t *terminal) Init() error {
	t.term.TPuts(t.c, t.term.EnterCA)
	t.term.TPuts(t.c, t.term.Clear)
	return nil
}

func (t *terminal) Fini() {
	t.term.TPuts(t.c, t.term.Clear)
	t.term.TPuts(t.c, t.term.ExitCA)
}

func (t *terminal) Clear() {
	for i := range t.current {
		t.current[i].c = ' '
		t.current[i].style = t.style
	}
}

func (t *terminal) Fill(c rune, st tcell.Style) {
	for i := range t.current {
		t.current[i].c = c
		t.current[i].style = st
	}
}

func (t *terminal) SetCell(x, y int, st tcell.Style, ch ...rune) {
	t.SetContent(x, y, ch[0], nil, st)
}

func (t *terminal) GetContent(x, y int) (mainc rune, combc []rune, style tcell.Style, width int) {
	cell := t.current[x+y*t.cols]
	return cell.c, nil, cell.style, 1
}

func (t *terminal) SetContent(x, y int, mainc rune, combc []rune, st tcell.Style) {
	t.current[x+y*t.cols].style = st
	t.current[x+y*t.cols].c = mainc
}

func (t *terminal) SetStyle(style tcell.Style) {
	t.style = style
}

func (t *terminal) ShowCursor(x int, y int) {
	t.term.TPuts(t.c, t.term.TParm(t.term.SetCursor, y, x))
	t.term.TPuts(t.c, t.term.ShowCursor)
}

func (t *terminal) HideCursor() {
	t.term.TPuts(t.c, t.term.HideCursor)
}

func (t *terminal) Size() (int, int) {
	return t.cols, t.rows
}

func (t *terminal) PollEvent() tcell.Event {
	return nil
}

func (t *terminal) PostEvent(ev tcell.Event) error {
	return tcell.ErrEventQFull
}

func (t *terminal) PostEventWait(ev tcell.Event) {
}

func (t *terminal) EnableMouse() {}

func (t *terminal) DisableMouse() {}

func (t *terminal) HasMouse() bool {
	return false
}

func (t *terminal) Colors() int {
	return 256
}

func (t *terminal) Show() {
	// TODO WHERE DO I EVEN START
	for x := 0; x < t.cols; x++ {
		for y := 0; y < t.rows; y++ {
			p := x + y*t.cols
			if t.current[p] != t.last[p] {
				t.term.TPuts(t.c, t.term.TParm(t.term.SetCursor, y, x))
				t.c.Write([]byte(string(t.current[p].c))) // ??
			}
		}
	}
	copy(t.last, t.current)
}

func (t *terminal) Sync() {
	for x := 0; x < t.cols; x++ {
		for y := 0; y < t.rows; y++ {
			t.term.TPuts(t.c, t.term.TParm(t.term.SetCursor, y, x))
			t.c.Write([]byte(string(t.current[x+y*t.cols].c))) // ??
		}
	}
	copy(t.last, t.current)
}

func (t *terminal) CharacterSet() string {
	return "UTF-8"
}

func (t *terminal) RegisterRuneFallback(r rune, subst string) {}

func (t *terminal) UnregisterRuneFallback(r rune) {}

func (t *terminal) CanDisplay(r rune, checkFallbacks bool) bool {
	return true
}

func (t *terminal) Resize(int, int, int, int) {}

func (t *terminal) HasKey(tcell.Key) bool {
	return true
}
