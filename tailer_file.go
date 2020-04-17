package tailer

import (
	"context"
	"io"
	"os"

	"github.com/sirupsen/logrus"
)

// NewFileTailer creates a new instance of the file-based Tailer
func NewFileTailer(ctx context.Context, fname string, out io.Writer) Tailer {
	var (
		bytesRead = int64(0)
		fd        *os.File
		st        os.FileInfo
		err       error
	)

	readFn := func(p []byte) (int, error) {
		if st, err = os.Stat(fname); os.IsNotExist(err) {
			logrus.Tracef("fileTailer: File %s does not yet exist", fname)
			return 0, io.EOF
		} else if fd == nil || err != nil || st.Size() < bytesRead {
			logrus.Tracef("fileTailer: File recycled -- reopening")
			bytesRead = 0
			if fd, err = os.Open(fname); err != nil {
				logrus.WithError(err).Warnf("fileTailer: Error opening file %s", fname)
				return 0, err
			}
		}
		return fd.Read(p)
	}

	return newTailer(ctx, readFn, out)
}
