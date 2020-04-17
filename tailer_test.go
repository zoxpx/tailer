package tailer

import (
	"bytes"
	"context"
	"io/ioutil"
	"log"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

var tmpFileName string

func write(t *testing.T, f *os.File, content string) *os.File {
	var err error

	if f == nil {
		f, err = os.OpenFile(tmpFileName, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0666)
		assert.NoError(t, err)
	}

	if len(content) > 0 {
		_, err = f.Write([]byte(content))
		assert.NoError(t, err)
	}
	return f
}

func TestExplicitStop(t *testing.T) {
	os.RemoveAll(tmpFileName)

	bb := bytes.Buffer{}
	tt := NewFileTailer(nil, tmpFileName, &bb).WithPoll(100 * time.Millisecond)
	tt.Start()

	f := write(t, nil, "Hello from TestExplicitStop\n")
	time.Sleep(110 * time.Millisecond)
	f.Write([]byte("wut?"))
	f.Close()
	time.Sleep(110 * time.Millisecond)
	tt.Stop()

	assert.Equal(t, "Hello from TestExplicitStop\nwut?", bb.String())
	time.Sleep(50 * time.Millisecond)
	assert.False(t, tt.IsRunning())
}

func TestContextStop(t *testing.T) {
	os.RemoveAll(tmpFileName)

	bb := bytes.Buffer{}
	ctx, ctxCancel := context.WithCancel(context.Background())
	tt := NewFileTailer(ctx, tmpFileName, &bb).WithPoll(100 * time.Millisecond).Start()

	write(t, nil, "Hello from TestContextStop\n").Close()

	time.Sleep(150 * time.Millisecond)
	ctxCancel()

	assert.Equal(t, "Hello from TestContextStop\n", bb.String())
	time.Sleep(50 * time.Millisecond)
	assert.False(t, tt.IsRunning())
}

func TestContextTimeoutStop(t *testing.T) {
	os.RemoveAll(tmpFileName)

	bb := bytes.Buffer{}
	ctx, _ := context.WithTimeout(context.Background(), 150*time.Millisecond)
	tt := NewFileTailer(ctx, tmpFileName, &bb).WithPoll(100 * time.Millisecond).Start()

	write(t, nil, "Hello from TestContextTimeoutStop\n").Close()

	time.Sleep(200 * time.Millisecond)

	assert.Equal(t, "Hello from TestContextTimeoutStop\n", bb.String())
	assert.False(t, tt.IsRunning())
}

func TestReaderTailer(t *testing.T) {
	bbI, bbO := bytes.Buffer{}, bytes.Buffer{}

	tt := NewReaderTailer(nil, &bbI, &bbO).WithPoll(100 * time.Millisecond)
	tt.Start()

	bbI.WriteString("Hello from TestReaderTailer\n")
	time.Sleep(150 * time.Millisecond)

	tt.Stop()
	assert.Equal(t, "Hello from TestReaderTailer\n", bbO.String())
	time.Sleep(50 * time.Millisecond)
	assert.False(t, tt.IsRunning())
}

func TestMain(m *testing.M) {
	tf, err := ioutil.TempFile("", "tailer")
	if err != nil {
		log.Fatal(err)
	}
	tmpFileName = tf.Name()
	tf.Close()
	defer os.RemoveAll(tmpFileName)

	os.Exit(m.Run())
}
