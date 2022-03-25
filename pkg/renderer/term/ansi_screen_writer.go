package term

import (
	"bytes"
	"fmt"
	"io"
	"strings"

	a "github.com/Azure/go-ansiterm"
)

// Writes terminal output to the screen, while allowing for per-line alterations.
type AnsiScreenWriter struct {
	parser           *a.AnsiParser
	currentRunOutput io.Writer
	reader           io.Reader
	stats            *ansiOutputStats
	flushBuffer      bytes.Buffer
	prefix           string
	beforeFlush      func()
	afterFlush       func()
}

type ansiOutputStats struct {
	cursor *cursorInfo
	lines  []lineInfo
}

type lineInfo struct {
	// We only write the line once there's printable characters. Until then, we store them here.
	unwrittenChars bytes.Buffer

	// True if any visible characters have been written to this line.
	hasVisibleChars bool

	// Indent has been written to this line. Safe to write without indenting.
	hasIndent bool
}

func NewAnsiScreenWriter(reader io.Reader, writer io.Writer) *AnsiScreenWriter {
	formatter := &AnsiScreenWriter{
		stats:            &ansiOutputStats{cursor: &cursorInfo{}},
		flushBuffer:      bytes.Buffer{},
		currentRunOutput: writer,
		reader:           reader,
	}
	formatter.parser = a.CreateParser("Ground", formatter)
	return formatter
}

func (f *AnsiScreenWriter) PrefixEachLine(prefix string) {
	f.prefix = prefix
}

func (f *AnsiScreenWriter) BeforeFlush(cb func()) {
	f.beforeFlush = cb
}

func (f *AnsiScreenWriter) AfterFlush(cb func()) {
	f.afterFlush = cb
}

func (f *AnsiScreenWriter) Process() {
	for {
		buffer := make([]byte, 4096)
		count, err := f.reader.Read(buffer)

		if err != nil {
			return
		}

		processed, err := f.parser.Parse(buffer[0:count])

		if err != nil {
			fmt.Printf("ERR: %s", err.Error())
			return
		}

		if processed != count {
			return
		}
	}
}

func (f *AnsiScreenWriter) allocLineStat(line int) {
	// Allocate line buffers to the cursor position.
	for i := len(f.stats.lines); i <= line+1; i++ {
		f.stats.lines = append(f.stats.lines, lineInfo{unwrittenChars: bytes.Buffer{}, hasVisibleChars: false})
	}
}

// Stores an invisible char into the line buffer
func (f *AnsiScreenWriter) storeUnwrittenString(s string) {
	line := f.stats.cursor.line

	f.allocLineStat(line)

	f.stats.lines[line].unwrittenChars.WriteString(s)

	f.stats.cursor.column++
}

// Print
func (f *AnsiScreenWriter) Print(b byte) error {
	line := f.stats.cursor.line
	f.allocLineStat(line)

	if !f.stats.lines[line].hasIndent {
		f.flushBuffer.WriteString(f.prefix)
		f.stats.lines[line].hasVisibleChars = true
		f.stats.lines[line].hasIndent = true
	}

	return f.flushBuffer.WriteByte(b)
}

// Execute C0 commands
func (f *AnsiScreenWriter) Execute(b byte) error {
	line := f.stats.cursor.line
	f.allocLineStat(line)

	f.flushBuffer.WriteByte(b)

	switch b {
	case 10: // CR+LF, or CR unix
		f.stats.cursor.line++
		f.stats.cursor.column = 0

		line := f.stats.cursor.line
		f.allocLineStat(line)
		f.stats.lines[line].hasIndent = false
	case 13: // CR
		f.stats.cursor.column = 0

		line := f.stats.cursor.line
		f.allocLineStat(line)
		f.stats.lines[line].hasIndent = false
	}

	return nil
}

// Cursor Up
func (f *AnsiScreenWriter) CUU(count int) error {
	f.stats.cursor.line -= count
	f.stats.cursor.column = 0

	if f.stats.cursor.line < 0 {
		f.stats.cursor.line = 0
	}

	line := f.stats.cursor.line
	f.allocLineStat(line)
	f.stats.lines[line].hasIndent = false

	f.flushBuffer.WriteString(fmt.Sprintf("\033[%dA", count))

	return nil
}

// CUrsor Down
func (f *AnsiScreenWriter) CUD(count int) error {
	f.stats.cursor.line += count

	return nil
}

// Cursor Forward
func (f *AnsiScreenWriter) CUF(count int) error {
	f.stats.cursor.column += count

	return nil
}

// Cursor Backward
func (f *AnsiScreenWriter) CUB(count int) error {
	f.stats.cursor.column -= count
	if f.stats.cursor.column < 0 {
		f.stats.cursor.column = 0
	}

	return nil
}

// Cursor to Next Line
func (f *AnsiScreenWriter) CNL(count int) error {
	f.stats.cursor.line += count
	f.stats.cursor.column = 0
	return nil
}

// Cursor to Previous Line
func (f *AnsiScreenWriter) CPL(count int) error {
	f.stats.cursor.line -= count
	f.stats.cursor.column = 0

	if f.stats.cursor.line < 0 {
		f.stats.cursor.line = 0
	}

	return nil
}

// Cursor Horizontal position Absolute
func (f *AnsiScreenWriter) CHA(pos int) error {
	f.stats.cursor.column = pos
	return nil
}

// Vertical line Position Absolute
func (f *AnsiScreenWriter) VPA(pos int) error {
	f.stats.cursor.line = pos
	return nil
}

// CUrsor Position
func (f *AnsiScreenWriter) CUP(x int, y int) error {
	f.stats.cursor.line = y
	f.stats.cursor.column = x
	return nil
}

// Horizontal and Vertical Position (depends on PUM)
func (f *AnsiScreenWriter) HVP(x int, y int) error {
	f.stats.cursor.line = y
	f.stats.cursor.column = x
	return nil

}

// Text Cursor Enable Mode
func (f *AnsiScreenWriter) DECTCEM(enable bool) error {
	return fmt.Errorf("not implemented: CursorEnable %+v", enable)

}

// Origin Mode
func (f *AnsiScreenWriter) DECOM(enable bool) error {
	return fmt.Errorf("not implemented: OriginMode %+v", enable)

}

// 132 Column Mode
func (f *AnsiScreenWriter) DECCOLM(enable bool) error {
	return fmt.Errorf("not implemented: 123ColumnMode %+v", enable)

}

// Erase in Display
func (f *AnsiScreenWriter) ED(count int) error {
	return fmt.Errorf("not implemented: EraseDisplay %+v", count)

}

// Erase in Line
func (f *AnsiScreenWriter) EL(count int) error {
	return fmt.Errorf("not implemented: EraseLine %+v", count)

}

// Insert Line
func (f *AnsiScreenWriter) IL(count int) error {
	return fmt.Errorf("not implemented: InsertLine %+v", count)

}

// Delete Line
func (f *AnsiScreenWriter) DL(count int) error {
	return fmt.Errorf("not implemented: DeleteLine %+v", count)

}

// Insert Character
func (f *AnsiScreenWriter) ICH(count int) error {
	return fmt.Errorf("not implemented: InsertChar %+v", count)

}

// Delete Character
func (f *AnsiScreenWriter) DCH(count int) error {
	return fmt.Errorf("not implemented: DeleteChar %+v", count)

}

// Set Graphics Rendition
func (f *AnsiScreenWriter) SGR(values []int) error {
	f.flushBuffer.WriteString("\033[")

	formatting := []string{}
	for _, v := range values {
		formatting = append(formatting, fmt.Sprintf("%d", v))
	}

	f.flushBuffer.WriteString(strings.Join(formatting, ";"))
	f.flushBuffer.WriteString("m")

	return nil
}

// Pan Down
func (f *AnsiScreenWriter) SU(count int) error {
	return fmt.Errorf("not implemented: PanDown %+v", count)

}

// Pan Up
func (f *AnsiScreenWriter) SD(count int) error {
	return fmt.Errorf("not implemented: PanUp %+v", count)

}

// Device Attributes
func (f *AnsiScreenWriter) DA(values []string) error {
	return fmt.Errorf("not implemented: DeviceAttrs %+v", values)

}

// Set Top and Bottom Margins
func (f *AnsiScreenWriter) DECSTBM(x, y int) error {
	return fmt.Errorf("not implemented: Margins %+v, %+v", x, y)

}

// Index
func (f *AnsiScreenWriter) IND() error {
	return fmt.Errorf("not implemented: Index")

}

// Reverse Index
func (f *AnsiScreenWriter) RI() error {
	return fmt.Errorf("not implemented: ReverseIndex")

}

// Flush updates from previous commands
func (f *AnsiScreenWriter) Flush() error {
	if f.beforeFlush != nil {
		f.beforeFlush()
	}

	f.currentRunOutput.Write(f.flushBuffer.Bytes())
	f.flushBuffer.Reset()

	type flushable interface {
		Flush() error
	}

	if flushWriter, ok := f.currentRunOutput.(flushable); ok {
		flushWriter.Flush()
	}

	if f.afterFlush != nil {
		f.afterFlush()
	}

	return nil
}
