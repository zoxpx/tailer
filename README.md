# File Tailer

Simple GO program that simulates `tail -F file`.

Can be controlled via Start/Stop:

```go
func main() {
	logrus.SetLevel(logrus.TraceLevel)
	ft := tailer.NewFileTailer("/tmp/tmp.xx", os.Stderr, nil).Start()
	go func() {
		time.Sleep(10 * time.Second)
		ft.Stop()
	}()
	time.Sleep(15 * time.Second)
}
```

Alternatively, control via context/cancellation:

```go
func main() {
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		time.Sleep(10 * time.Second)
		cancel()
	}()

	logrus.SetLevel(logrus.TraceLevel)
	tailer.NewFileTailer("/tmp/tmp.xx", os.Stderr, ctx).Start()
	time.Sleep(15 * time.Second)
}
```

Or control via context Timeout:

```go

func main() {
	ctx, _ := context.WithTimeout(context.Background(), 10 * time.Second)

	logrus.SetLevel(logrus.TraceLevel)
	tailer.NewFileTailer("/tmp/tmp.xx", os.Stderr, ctx).Start()
	time.Sleep(15 * time.Second)
}
```
