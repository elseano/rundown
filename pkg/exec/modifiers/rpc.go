package modifiers

import (
	"bytes"
	"io"
	"os"

	"github.com/Azure/go-ansiterm"
)

type CommandHandler func(s string)

type RpcStream struct {
	ExecutionModifier
	writer  *os.File
	Reader  *os.File
	handler CommandHandler
}

func NewRpcStream(handler CommandHandler) *RpcStream {
	reader, writer, _ := os.Pipe()

	return &RpcStream{
		ExecutionModifier: &NullModifier{},
		Reader:            reader,
		writer:            writer,
		handler:           handler,
	}
}

func (m *RpcStream) GetStdout() []io.Writer {
	go func() {
		handler := &streamHandler{handler: m.handler}
		parser := ansiterm.CreateParser("Ground", handler)

		for {
			buf := make([]byte, 4096)
			_, err := m.Reader.Read(buf)

			if err != nil {
				return
			}

			parser.Parse(buf)
		}
	}()

	return []io.Writer{m.writer}
}

func (m *RpcStream) GetResult(int) []ModifierResult {
	m.writer.Close()
	return []ModifierResult{}
}

type streamHandler struct {
	oscBuffer bytes.Buffer
	handler   func(string)
}

func (h *streamHandler) OSC(b []byte) error {
	_, err := h.oscBuffer.Write(b)
	return err
}

func (h *streamHandler) Print(b byte) error {
	return nil
}
func (h *streamHandler) Execute(b byte) error {
	return nil
}
func (h *streamHandler) CUU(int) error {
	return nil
}
func (h *streamHandler) CUD(int) error {
	return nil
}
func (h *streamHandler) CUF(int) error {
	return nil
}
func (h *streamHandler) CUB(int) error {
	return nil
}
func (h *streamHandler) CNL(int) error {
	return nil
}
func (h *streamHandler) CPL(int) error {
	return nil
}
func (h *streamHandler) CHA(int) error {
	return nil
}
func (h *streamHandler) VPA(int) error {
	return nil
}
func (h *streamHandler) CUP(int, int) error {
	return nil
}
func (h *streamHandler) HVP(int, int) error {
	return nil
}
func (h *streamHandler) DECTCEM(bool) error {
	return nil
}
func (h *streamHandler) DECOM(bool) error {
	return nil
}
func (h *streamHandler) DECCOLM(bool) error {
	return nil
}
func (h *streamHandler) ED(int) error {
	return nil
}
func (h *streamHandler) EL(int) error {
	return nil
}
func (h *streamHandler) IL(int) error {
	return nil
}
func (h *streamHandler) DL(int) error {
	return nil
}
func (h *streamHandler) ICH(int) error {
	return nil
}
func (h *streamHandler) DCH(int) error {
	return nil
}
func (h *streamHandler) SGR([]int) error {
	return nil
}
func (h *streamHandler) SU(int) error {
	return nil
}
func (h *streamHandler) SD(int) error {
	return nil
}
func (h *streamHandler) DA([]string) error {
	return nil
}
func (h *streamHandler) DECSTBM(int, int) error {
	return nil
}
func (h *streamHandler) IND() error {
	return nil
}
func (h *streamHandler) RI() error {
	return nil
}
func (h *streamHandler) Flush() error {
	result := h.oscBuffer.String()
	if result != "" {
		h.handler(result)
		h.oscBuffer.Reset()
	}

	return nil
}
