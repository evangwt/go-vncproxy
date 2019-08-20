package vncproxy

import (
	"fmt"
	"log"
)

const (
	InfoLevel  uint32 = 0x1 << 0
	DebugLevel uint32 = InfoLevel | 0x1<<1
)

type logger struct {
	level uint32
}

func NewLogger(level uint32) *logger {
	return &logger{
		level: level,
	}
}

func (l *logger) Info(msg ...interface{}) {
	if l.level&InfoLevel > 0 {
		log.Printf("[vncproxy][info] %v", fmt.Sprint(msg...))
	}
}

func (l *logger) Infof(format string, msg ...interface{}) {
	if l.level&InfoLevel > 0 {
		log.Printf("[vncproxy][info] %v", fmt.Sprintf(format, msg...))
	}
}

func (l *logger) Debug(msg ...interface{}) {
	if l.level&DebugLevel > 0 {
		log.Printf("[vncproxy][debug] %v", fmt.Sprint(msg...))
	}
}

func (l *logger) Debugf(format string, msg ...interface{}) {
	if l.level&DebugLevel > 0 {
		log.Printf("[vncproxy][debug] %v", fmt.Sprintf(format, msg...))
	}
}
