package main

import (
	"bytes"
	"context"
	"fmt"
	"image"
	"io"

	"github.com/mum4k/termdash/cell"
	"github.com/xo/terminfo"
)

type terminal struct {
	c          io.ReadWriter
	cols, rows int
	term       *terminfo.Terminfo
	outbuf     *bytes.Buffer
}

func (t *terminal) Init(cols, rows int, term string) error {
	t.cols = cols
	t.rows = rows
	var err error
	t.term, err = terminfo.Load(term)
	if err != nil {
		return err
	}
	t.outbuf = new(bytes.Buffer)
	t.term.Fprintf(t.outbuf, terminfo.EnterCaMode)
	t.term.Fprintf(t.outbuf, terminfo.ClearScreen)
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

// interface for termdash.terminal.Terminal

func (t *terminal) Size() image.Point {
	return image.Point{
		X: t.cols,
		Y: t.rows,
	}
}

// Clear clears the content of the internal back buffer, resetting all
// cells to their default content and attributes. Sets the provided options
// on all the cell.
func (t *terminal) Clear(opts ...cell.Option) error {
	return nil
}

// Flush flushes the internal back buffer to the terminal.
func (t *terminal) Flush() error {
	t.c.Write(t.outbuf.Bytes())
	t.outbuf.Reset()
	return nil
}

// SetCursor sets the position of the cursor.
func (t *terminal) SetCursor(p image.Point) {}

// HideCursos hides the cursor.
func (t *terminal) HideCursor() {
	t.term.Fprintf(t.outbuf, terminfo.CursorInvisible)
	t.Flush()
}

// SetCell sets the value of the specified cell to the provided rune.
// Use the options to specify which attributes to modify, if an attribute
// option isn't specified, the attribute retains its previous value.
func (t *terminal) SetCell(p image.Point, r rune, opts ...cell.Option) error {
	return nil
}

// Event waits for the next event and returns it.
// This call blocks until the next event or cancellation of the context.
// Returns nil when the context gets canceled.
func (t *terminal) Event(ctx context.Context) {}
