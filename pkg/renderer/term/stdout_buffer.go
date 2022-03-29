package term

import (
	"fmt"
	"io"
	"strings"

	a "github.com/Azure/go-ansiterm"
)

// Buffers an output stream. Consecutive flushes should reposition cursor to the original cursor point.
type StdoutBuffer struct {
	parser           *a.AnsiParser
	currentRunOutput io.Writer
	buffer           *outputBuffer
	anythingToFlush  bool
	LastFlushLines   int
	flushChan        chan bool
}

type outputBuffer struct {
	cursor *cursorInfo
	lines  []lineBuffer
}

type lineBuffer struct {
	contents        strings.Builder
	hasVisibleChars bool
}

type cursorInfo struct {
	line   int
	column int
}

func NewStdoutBuffer() *StdoutBuffer {
	formatter := &StdoutBuffer{
		buffer: &outputBuffer{cursor: &cursorInfo{}},
	}
	formatter.parser = a.CreateParser("Ground", formatter)
	return formatter
}

func (f *StdoutBuffer) SubscribeToFlush() chan bool {
	f.flushChan = make(chan bool)
	return f.flushChan
}

func (f *StdoutBuffer) Process(inStream io.Reader) {
	for {
		buffer := make([]byte, 4096)
		count, err := inStream.Read(buffer)

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

func (b outputBuffer) String() string {
	output := strings.Builder{}
	for _, line := range b.lines {
		output.WriteString(fmt.Sprintf("%s\r\n", line.contents.String()))
	}

	return output.String()
}

func (f *StdoutBuffer) allocBuffer(line int) {
	// Allocate line buffers to the cursor position.
	for i := len(f.buffer.lines); i <= line+1; i++ {
		f.buffer.lines = append(f.buffer.lines, lineBuffer{hasVisibleChars: false})
	}
}

func (f *StdoutBuffer) writeToBuffer(b byte) {
	line := f.buffer.cursor.line

	f.allocBuffer(line)

	f.buffer.lines[line].contents.WriteByte(b)
	f.buffer.lines[line].hasVisibleChars = true
	f.anythingToFlush = true

	f.buffer.cursor.column++
}

func (f *StdoutBuffer) writeStringToBuffer(str string) {
	line := f.buffer.cursor.line

	f.allocBuffer(line)

	f.buffer.lines[line].contents.WriteString(str)
	f.buffer.lines[line].hasVisibleChars = true
	f.anythingToFlush = true

	f.buffer.cursor.column += len(str)
}

// Print
func (f *StdoutBuffer) Print(b byte) error {
	f.writeToBuffer(b)
	return nil
}

// Execute C0 commands
func (f *StdoutBuffer) Execute(b byte) error {
	switch b {
	case 10: // CR+LF, or CR unix
		f.buffer.cursor.line++
		f.buffer.cursor.column = 0
	case 13: // CR
		f.buffer.cursor.column = 0
	default:
		f.writeStringToBuffer(fmt.Sprintf("<EXEC %d>", b))
		// return fmt.Errorf("not implemented: Execute %d", b)
	}

	return nil
}

// Cursor Up
func (f *StdoutBuffer) CUU(count int) error {
	f.buffer.cursor.line -= count
	if f.buffer.cursor.line < 0 {
		f.buffer.cursor.line = 0
	}

	return nil
}

// CUrsor Down
func (f *StdoutBuffer) CUD(count int) error {
	f.buffer.cursor.line += count

	return nil
}

// Cursor Forward
func (f *StdoutBuffer) CUF(count int) error {
	f.buffer.cursor.column += count

	return nil
}

// Cursor Backward
func (f *StdoutBuffer) CUB(count int) error {
	f.buffer.cursor.column -= count
	if f.buffer.cursor.column < 0 {
		f.buffer.cursor.column = 0
	}

	return nil
}

// Cursor to Next Line
func (f *StdoutBuffer) CNL(count int) error {
	f.buffer.cursor.line += count
	f.buffer.cursor.column = 0
	return nil
}

// Cursor to Previous Line
func (f *StdoutBuffer) CPL(count int) error {
	f.buffer.cursor.line -= count
	f.buffer.cursor.column = 0

	if f.buffer.cursor.line < 0 {
		f.buffer.cursor.line = 0
	}

	return nil
}

// Cursor Horizontal position Absolute
func (f *StdoutBuffer) CHA(pos int) error {
	f.buffer.cursor.column = pos
	return nil
}

// Vertical line Position Absolute
func (f *StdoutBuffer) VPA(pos int) error {
	f.buffer.cursor.line = pos
	return nil
}

// CUrsor Position
func (f *StdoutBuffer) CUP(x int, y int) error {
	f.buffer.cursor.line = y
	f.buffer.cursor.column = x
	return nil
}

// Horizontal and Vertical Position (depends on PUM)
func (f *StdoutBuffer) HVP(x int, y int) error {
	f.buffer.cursor.line = y
	f.buffer.cursor.column = x
	return nil

}

// Text Cursor Enable Mode
func (f *StdoutBuffer) DECTCEM(enable bool) error {
	return fmt.Errorf("not implemented: CursorEnable %+v", enable)

}

// Origin Mode
func (f *StdoutBuffer) DECOM(enable bool) error {
	return fmt.Errorf("not implemented: OriginMode %+v", enable)

}

// 132 Column Mode
func (f *StdoutBuffer) DECCOLM(enable bool) error {
	return fmt.Errorf("not implemented: 123ColumnMode %+v", enable)

}

// Erase in Display
func (f *StdoutBuffer) ED(count int) error {
	return fmt.Errorf("not implemented: EraseDisplay %+v", count)

}

// Erase in Line
func (f *StdoutBuffer) EL(count int) error {
	return fmt.Errorf("not implemented: EraseLine %+v", count)

}

// Insert Line
func (f *StdoutBuffer) IL(count int) error {
	return fmt.Errorf("not implemented: InsertLine %+v", count)

}

// Delete Line
func (f *StdoutBuffer) DL(count int) error {
	return fmt.Errorf("not implemented: DeleteLine %+v", count)

}

// Insert Character
func (f *StdoutBuffer) ICH(count int) error {
	return fmt.Errorf("not implemented: InsertChar %+v", count)

}

// Delete Character
func (f *StdoutBuffer) DCH(count int) error {
	return fmt.Errorf("not implemented: DeleteChar %+v", count)

}

// Set Graphics Rendition
func (f *StdoutBuffer) SGR(values []int) error {
	f.writeStringToBuffer("\033[")

	formatting := []string{}
	for _, v := range values {
		formatting = append(formatting, fmt.Sprintf("%d", v))
	}

	f.writeStringToBuffer(strings.Join(formatting, ";"))
	f.writeToBuffer('m')

	return nil
}

// Pan Down
func (f *StdoutBuffer) SU(count int) error {
	return fmt.Errorf("not implemented: PanDown %+v", count)

}

// Pan Up
func (f *StdoutBuffer) SD(count int) error {
	return fmt.Errorf("not implemented: PanUp %+v", count)

}

// Device Attributes
func (f *StdoutBuffer) DA(values []string) error {
	return fmt.Errorf("not implemented: DeviceAttrs %+v", values)

}

// Set Top and Bottom Margins
func (f *StdoutBuffer) DECSTBM(x, y int) error {
	return fmt.Errorf("not implemented: Margins %+v, %+v", x, y)

}

// Index
func (f *StdoutBuffer) IND() error {
	return fmt.Errorf("not implemented: Index")

}

// Reverse Index
func (f *StdoutBuffer) RI() error {
	return fmt.Errorf("not implemented: ReverseIndex")

}

// Flush updates from previous commands
func (f *StdoutBuffer) Flush() error {
	if f.flushChan != nil {
		f.flushChan <- true
	}

	return nil
}

func (f *StdoutBuffer) AnythingToFlush() bool {
	return f.anythingToFlush
}

func (f *StdoutBuffer) String() (string, int) {
	return f.buffer.String(), len(f.buffer.lines)
}
