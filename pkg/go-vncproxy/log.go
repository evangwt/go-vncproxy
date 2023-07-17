package vncproxy

import (
	"fmt"
	"log"
)

const (
	InfoFlag  uint32 = 0x1 << 0
	DebugFlag uint32 = 0x1 << 1

	InfoLevel  uint32 = InfoFlag
	DebugLevel uint32 = InfoFlag | DebugFlag
)

type Logger interface {
	Infof(format string, v ...interface{})
	Debugf(format string, v ...interface{})
}

type logger struct {
	level  uint32
	logger Logger
}

func NewLogger(level uint32, instance Logger) *logger {
	return &logger{
		level:  level,
		logger: instance,
	}
}

func (l *logger) print(level string, msg ...interface{}) {
	if l.logger != nil {
		if level == "debug" {
			l.logger.Debugf("[vncproxy] %v", fmt.Sprint(msg...))
		} else {
			l.logger.Infof("[vncproxy] %v", fmt.Sprint(msg...))
		}
	} else {
		log.Printf("[vncproxy][%v] %v", level, fmt.Sprint(msg...))
	}
}

func (l *logger) Info(msg ...interface{}) {
	if l.level&InfoFlag > 0 {
		l.print("info", msg...)
	}
}

func (l *logger) Infof(format string, msg ...interface{}) {
	if l.level&InfoFlag > 0 {
		l.print("info", fmt.Sprintf(format, msg...))
	}
}

func (l *logger) Debug(msg ...interface{}) {
	if l.level&DebugFlag > 0 {
		l.print("debug", msg...)
	}
}

func (l *logger) Debugf(format string, msg ...interface{}) {
	if l.level&DebugFlag > 0 {
		l.print("debug", fmt.Sprintf(format, msg...))
	}
}
