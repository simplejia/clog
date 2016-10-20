package clog

import "testing"

func TestClog(t *testing.T) {
	Init("demo", "", 15, 3)
	Debug("debug msg")
	Warn("warn msg")
	Error("error msg")
	Info("info msg")
}
