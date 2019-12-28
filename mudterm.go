package main

import (
	"fmt"
	"io"

	"github.com/gdamore/tcell"
	"github.com/xo/terminfo"
)

type cell struct {
	c     rune
	style tcell.Style
}

type terminal struct {
	c                io.ReadWriter
	cols, rows       int
	term             *terminfo.Terminfo
	current, updated []cell
	style            tcell.Style
}

func (t *terminal) SetInfo(cols, rows int, term string) error {
	t.cols = cols
	t.rows = rows
	var err error
	t.term, err = terminfo.Load(term)
	if err != nil {
		return err
	}
	t.current = make([]cell, cols*rows)
	t.updated = make([]cell, cols*rows)
	t.style = tcell.StyleDefault
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
	t.term.Fprintf(t.c, terminfo.EnterCaMode)
	t.term.Fprintf(t.c, terminfo.ClearScreen)
	return nil
}

func (t *terminal) Fini() {
	t.term.Fprintf(t.c, terminfo.ClearScreen)
	t.term.Fprintf(t.c, terminfo.ExitCaMode)
}

func (t *terminal) Clear() {
	for i := range t.updated {
		t.updated[i].c = ' '
		t.updated[i].style = t.style
	}
}

func (t *terminal) Fill(c rune, st tcell.Style) {
	for i := range t.updated {
		t.updated[i].c = c
		t.updated[i].style = st
	}
}

func (t *terminal) SetCell(x, y int, st tcell.Style, ch ...rune) {
	t.SetContent(x, y, ch[0], nil, st)
}

func (t *terminal) GetContent(x, y int) (mainc rune, combc []rune, style tcell.Style, width int) {
	cell := t.updated[x+y*t.cols]
	return cell.c, nil, cell.style, 1
}

func (t *terminal) SetContent(x, y int, mainc rune, combc []rune, st tcell.Style) {
	t.updated[x+y*t.cols].style = st
	t.updated[x+y*t.cols].c = mainc
}

func (t *terminal) SetStyle(style tcell.Style) {
	t.style = style
}

func (t *terminal) ShowCursor(x int, y int) {
	io.WriteString(t.c, t.term.Goto(y, x))
	t.term.Fprintf(t.c, terminfo.CursorVisible)
}

func (t *terminal) HideCursor() {
	t.term.Fprintf(t.c, terminfo.CursorInvisible)
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

func (t *terminal) Show() {}

func (t *terminal) Sync() {}

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
