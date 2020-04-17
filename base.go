package tailer

import (
	"context"
	"io"
	"os"
	"time"

	"github.com/sirupsen/logrus"
)

// Tailer simple interface to control the tail-follow of the file or io.Reader
type Tailer interface {
	Start() Tailer
	Stop()
	IsRunning() bool
	WithPoll(time.Duration) Tailer
}

type tailer struct {
	readFn    func([]byte) (int, error)
	out       io.Writer
	isRunning bool
	ctx       context.Context
	cancel    context.CancelFunc
	poll      time.Duration
}

// newTailer returns a new instance of Tailer
func newTailer(ctx context.Context, readFn func([]byte) (int, error), out io.Writer) Tailer {
	if ctx == nil {
		ctx = context.Background()
	}
	ctx2, cancel := context.WithCancel(ctx)
	if out == nil {
		out = os.Stderr
	}

	return &tailer{
		readFn: readFn,
		out:    out,
		ctx:    ctx2,
		cancel: cancel,
		poll:   time.Second,
	}
}

func (f *tailer) loop() error {
	var (
		buf = make([]byte, 2048)
		n   int
		err error
	)

	f.isRunning = true
	defer func() { f.isRunning = false }()

	for {
		if n, err = f.readFn(buf); n > 0 {
			logrus.Tracef("tailer: Read %d bytes", n)
			f.out.Write(buf[0:n])
		}

		if n == 0 || err == io.EOF {
			goto labSleep
		} else if err != nil {
			logrus.WithError(err).Warnf("tailer: Error reading")
			return err
		}

	labSleep:
		select {
		case <-time.After(f.poll):
			// nop
		case <-f.ctx.Done():
			logrus.Debugf("tailer loop done.")
			return f.ctx.Err()
		}
	}
}

func (f *tailer) Start() Tailer {
	go f.loop()
	return f
}

func (f *tailer) Stop() {
	f.cancel()
}

func (f *tailer) IsRunning() bool {
	return f.isRunning
}

func (f *tailer) WithPoll(dur time.Duration) Tailer {
	f.poll = dur
	return f
}
