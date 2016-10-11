package util

import (
	"io"

	"github.com/golang/glog"
)

// GetLogLevel returns the current glog log level
func GetLogLevel() (level int) {
	for level = 5; level >= 0; level-- {
		if glog.V(glog.Level(level)) == true {
			break
		}
	}
	return
}

// NewGLogWriterV returns a new Writer that delegates to `glog.Info` at the
// desired level of verbosity
func NewGLogWriterV(level int) io.Writer {
	return &gLogWriter{
		level: glog.Level(level),
	}
}

// gLogWriter is a Writer that writes by delegating to `glog.Info`
type gLogWriter struct {
	// level is the default level to log at
	level glog.Level
}

func (w *gLogWriter) Write(p []byte) (n int, err error) {
	glog.V(w.level).InfoDepth(2, string(p))

	return len(p), nil
}
