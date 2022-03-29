package term

import "io"

type FlushingWriter struct {
	writer io.Writer
}

func NewFlushingWriter(w io.Writer) *FlushingWriter {
	return &FlushingWriter{w}
}

func (f *FlushingWriter) Write(b []byte) (int, error) {
	count, err := f.writer.Write(b)

	if ff, ok := f.writer.(interface{ Flush() error }); ok && err == nil {
		return count, ff.Flush()
	}

	return count, err
}
