package core

import (
	"io"
	"os"
	"path/filepath"
	"strconv"

	"github.com/itchyny/bed/buffer"
	"github.com/itchyny/bed/util"
)

// Window represents an editor window.
type Window struct {
	buffer   *buffer.Buffer
	name     string
	basename string
	height   int64
	width    int64
	offset   int64
	cursor   int64
	length   int64
	stack    []position
	mode     Mode
}

type position struct {
	cursor int64
	offset int64
}

// NewWindow creates a new editor window.
func NewWindow(name string, width int64) (*Window, error) {
	file, err := os.Open(name)
	if err != nil {
		return nil, err
	}
	buffer := buffer.NewBuffer(file)
	length, err := buffer.Len()
	if err != nil {
		return nil, err
	}
	return &Window{
		buffer:   buffer,
		name:     name,
		basename: filepath.Base(name),
		width:    width,
		length:   length,
		mode:     ModeNormal,
	}, nil
}

func (w *Window) readBytes(pos int64, len int) (int, []byte, error) {
	bytes := make([]byte, len)
	_, err := w.buffer.Seek(pos, io.SeekStart)
	if err != nil {
		return 0, bytes, err
	}
	n, err := w.buffer.Read(bytes)
	if err != nil && err != io.EOF {
		return 0, bytes, err
	}
	return n, bytes, nil
}

// State returns the current state of the buffer.
func (w *Window) State() (State, error) {
	n, bytes, err := w.readBytes(w.offset, int(w.height*w.width))
	if err != nil {
		return State{}, err
	}
	return State{
		Name:   w.basename,
		Width:  int(w.width),
		Offset: w.offset,
		Cursor: w.cursor,
		Bytes:  bytes,
		Size:   n,
		Length: w.length,
		Mode:   w.mode,
	}, nil
}

// Close the window.
func (w *Window) Close() error {
	if w.buffer == nil {
		return nil
	}
	return w.buffer.Close()
}

func (w *Window) cursorUp(count int64) {
	w.cursor -= util.MinInt64(util.MaxInt64(count, 1), w.cursor/w.width) * w.width
	if w.cursor < w.offset {
		w.offset = w.cursor / w.width * w.width
	}
}

func (w *Window) cursorDown(count int64) {
	w.cursor += util.MinInt64(
		util.MinInt64(util.MaxInt64(count, 1), (util.MaxInt64(w.length, 1)-1)/w.width-w.cursor/w.width)*w.width,
		util.MaxInt64(w.length, 1)-1-w.cursor)
	if w.cursor >= w.offset+w.height*w.width {
		w.offset = (w.cursor - w.height*w.width + w.width) / w.width * w.width
	}
}

func (w *Window) cursorLeft(count int64) {
	w.cursor -= util.MinInt64(util.MaxInt64(count, 1), w.cursor%w.width)
}

func (w *Window) cursorRight(count int64) {
	w.cursor += util.MinInt64(util.MinInt64(util.MaxInt64(count, 1), w.width-1-w.cursor%w.width), util.MaxInt64(w.length, 1)-1-w.cursor)
}

func (w *Window) cursorPrev(count int64) {
	w.cursor -= util.MinInt64(util.MaxInt64(count, 1), w.cursor)
	if w.cursor < w.offset {
		w.offset = w.cursor / w.width * w.width
	}
}

func (w *Window) cursorNext(count int64) {
	w.cursor += util.MinInt64(util.MaxInt64(count, 1), util.MaxInt64(w.length, 1)-1-w.cursor)
	if w.cursor >= w.offset+w.height*w.width {
		w.offset = (w.cursor - w.height*w.width + w.width) / w.width * w.width
	}
}

func (w *Window) cursorHead(_ int64) {
	w.cursor -= w.cursor % w.width
}

func (w *Window) cursorEnd(count int64) {
	w.cursor = util.MinInt64((w.cursor/w.width+util.MaxInt64(count, 1))*w.width-1, util.MaxInt64(w.length, 1)-1)
	if w.cursor >= w.offset+w.height*w.width {
		w.offset = (w.cursor - w.height*w.width + w.width) / w.width * w.width
	}
}

func (w *Window) scrollUp(count int64) {
	w.offset -= util.MinInt64(util.MaxInt64(count, 1), w.offset/w.width) * w.width
	if w.cursor >= w.offset+w.height*w.width {
		w.cursor -= ((w.cursor-w.offset-w.height*w.width)/w.width + 1) * w.width
	}
}

func (w *Window) scrollDown(count int64) {
	h := (util.MaxInt64(w.length, 1)+w.width-1)/w.width - w.height
	w.offset += util.MinInt64(util.MaxInt64(count, 1), h-w.offset/w.width) * w.width
	if w.cursor < w.offset {
		w.cursor += util.MinInt64((w.offset-w.cursor+w.width-1)/w.width*w.width, util.MaxInt64(w.length, 1)-1-w.cursor)
	}
}

func (w *Window) pageUp() {
	w.offset = util.MaxInt64(w.offset-(w.height-2)*w.width, 0)
	if w.offset == 0 {
		w.cursor = 0
	} else if w.cursor >= w.offset+w.height*w.width {
		w.cursor = w.offset + (w.height-1)*w.width
	}
}

func (w *Window) pageDown() {
	offset := util.MaxInt64(((w.length+w.width-1)/w.width-w.height)*w.width, 0)
	w.offset = util.MinInt64(w.offset+(w.height-2)*w.width, offset)
	if w.cursor < w.offset {
		w.cursor = w.offset
	} else if w.offset == offset {
		w.cursor = ((util.MaxInt64(w.length, 1)+w.width-1)/w.width - 1) * w.width
	}
}

func (w *Window) pageUpHalf() {
	w.offset = util.MaxInt64(w.offset-util.MaxInt64(w.height/2, 1)*w.width, 0)
	if w.offset == 0 {
		w.cursor = 0
	} else if w.cursor >= w.offset+w.height*w.width {
		w.cursor = w.offset + (w.height-1)*w.width
	}
}

func (w *Window) pageDownHalf() {
	offset := util.MaxInt64(((w.length+w.width-1)/w.width-w.height)*w.width, 0)
	w.offset = util.MinInt64(w.offset+util.MaxInt64(w.height/2, 1)*w.width, offset)
	if w.cursor < w.offset {
		w.cursor = w.offset
	} else if w.offset == offset {
		w.cursor = ((util.MaxInt64(w.length, 1)+w.width-1)/w.width - 1) * w.width
	}
}

func (w *Window) pageTop() {
	w.offset = 0
	w.cursor = 0
}

func (w *Window) pageEnd() {
	w.offset = util.MaxInt64(((w.length+w.width-1)/w.width-w.height)*w.width, 0)
	w.cursor = ((util.MaxInt64(w.length, 1)+w.width-1)/w.width - 1) * w.width
}

func isDigit(b byte) bool {
	return '\x30' <= b && b <= '\x39'
}

func isWhite(b byte) bool {
	return b == '\x00' || b == '\x09' || b == '\x0a' || b == '\x0d' || b == '\x20'
}

func (w *Window) jumpTo() {
	s := 50
	_, bytes, err := w.readBytes(util.MaxInt64(w.cursor-int64(s), 0), 2*s)
	if err != nil {
		return
	}
	var i, j int
	for i = s; i < 2*s && isWhite(bytes[i]); i++ {
	}
	if i == 2*s || !isDigit(bytes[i]) {
		return
	}
	for ; 0 < i && isDigit(bytes[i-1]); i-- {
	}
	for j = i; j < 2*s && isDigit(bytes[j]); j++ {
	}
	if j == 2*s {
		return
	}
	offset, _ := strconv.ParseInt(string(bytes[i:j]), 10, 64)
	if offset <= 0 || w.length <= offset {
		return
	}
	w.stack = append(w.stack, position{w.cursor, w.offset})
	w.cursor = offset
	w.offset = util.MaxInt64(offset-offset%w.width-util.MaxInt64(w.height/3, 0)*w.width, 0)
}

func (w *Window) jumpBack() {
	if len(w.stack) == 0 {
		return
	}
	w.cursor = w.stack[len(w.stack)-1].cursor
	w.offset = w.stack[len(w.stack)-1].offset
	w.stack = w.stack[:len(w.stack)-1]
}

func (w *Window) startInsert() {
	w.mode = ModeInsert
}

func (w *Window) exitInsert() {
	w.mode = ModeNormal
}
