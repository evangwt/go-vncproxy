package vncproxy

import (
	"fmt"
	"log"
	"testing"
)

func TestNewLogger(t *testing.T) {
	l := NewLogger(DebugLevel, nil)
	l.Infof("info message")
	l.Debugf("debug message")
}

type myLogger struct {
	Format string
	Msg    string
}

func (l *myLogger) Infof(format string, v ...interface{}) {
	l.Format = format
	l.Msg = fmt.Sprintf("%v", v...)
	log.Printf(format, v...)
}

func (l *myLogger) Debugf(format string, v ...interface{}) {
	l.Format = format
	l.Msg = fmt.Sprintf("%s", v...)
	log.Printf(format, v...)
}

func TestNewCustomLogger(t *testing.T) {
	_logger := &myLogger{}
	l := NewLogger(DebugLevel, _logger)

	fn1 := []func(...interface{}){l.Info, l.Debug}
	for i, f := range fn1 {
		msg := fmt.Sprintf("some messages %d", i)
		f(msg)
		if _logger.Format != "[vncproxy] %v" {
			t.Fatal("incorrect logger format")
		}
		if _logger.Msg != msg {
			t.Fatal("incorrect logger msg")
		}
	}

	fn2 := []func(string, ...interface{}){l.Infof, l.Debugf}
	format := "prefix %v"
	for i, f := range fn2 {
		msg := fmt.Sprintf("some messages %d", i)
		f(format, msg)
		if _logger.Format != "[vncproxy] %v" {
			t.Fatal("incorrect logger format")
		}
		if _logger.Msg != fmt.Sprintf(format, msg) {
			t.Fatal("incorrect logger msg")
		}
	}
}
