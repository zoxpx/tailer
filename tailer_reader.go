package tailer

import (
	"context"
	"io"
)

// NewReaderTailer creates a new instance of the reader-based Tailer
func NewReaderTailer(ctx context.Context, in io.Reader, out io.Writer) Tailer {
	return newTailer(ctx, in.Read, out)
}
