package tailer

import (
	"context"
	"github.com/sirupsen/logrus"
	"io"
	"os"
	"time"
)

// FileTailer simple interface to control the tail-follow of the file
type FileTailer interface {
	Start() FileTailer
	Stop()
	IsRunning() bool
}

type fileTailer struct {
	fname     string
	out       io.Writer
	isRunning bool
	ctx       context.Context
	cancel    context.CancelFunc
}

// NewFileTailer creates a new instance of the FileTailer
func NewFileTailer(fname string, out io.Writer, ctx context.Context) FileTailer {
	if out == nil {
		out = os.Stderr
	}
	if ctx == nil {
		ctx = context.Background()
	}
	ctx2, cancel := context.WithCancel(ctx)
	return &fileTailer{
		fname:  fname,
		out:    out,
		ctx:    ctx2,
		cancel: cancel,
	}
}

func (f *fileTailer) loop() error {
	var (
		fd        *os.File
		buf       = make([]byte, 2048)
		bytesRead = int64(0)
		n         int
		st        os.FileInfo
		err       error
	)

	f.isRunning = true
	defer func() { f.isRunning = false }()

	for {
		if st, err = os.Stat(f.fname); os.IsNotExist(err) {
			logrus.Tracef("fileTailer: File %s does not yet exist", f.fname)
			goto labSleep
		} else if fd == nil || err != nil || st.Size() < bytesRead {
			logrus.Tracef("fileTailer: File recycled -- reopening")
			bytesRead = 0
			if fd, err = os.Open(f.fname); err != nil {
				logrus.WithError(err).Warnf("fileTailer: Error opening file %s", f.fname)
				return err
			}
		}

		if n, err = fd.Read(buf); err == nil {
			logrus.Tracef("fileTailer: Read %d bytes", n)
			f.out.Write(buf[0:n])
			bytesRead += int64(n)
		} else if err == io.EOF {
			goto labSleep
		} else {
			logrus.WithError(err).Warnf("fileTailer: Error reading file %s", f.fname)
			return err
		}

	labSleep:
		select {
		case <-time.After(time.Second):
			// nop
		case <-f.ctx.Done():
			logrus.Debugf("fileTailer loop done.")
			return f.ctx.Err()
		}
	}
}

func (f *fileTailer) Start() FileTailer {
	go f.loop()
	return f
}

func (f *fileTailer) Stop() {
	f.cancel()
}

func (f *fileTailer) IsRunning() bool {
	return f.isRunning
}