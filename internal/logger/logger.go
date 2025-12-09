package logger

import (
	"fmt"
	"io"
	"os"
	"runtime/debug"
	"sync"
	"time"
)

type Level int8

const (
	LevelInfo Level = iota
	LevelError
	LevelFatal
	LevelOff
)

func (l Level) String() string {
	switch l {
	case LevelInfo:
		return "INFO"
	case LevelError:
		return "ERORR"
	case LevelFatal:
		return "FATAL"
	default:
		return ""
	}
}

type Logger struct {
	out      io.Writer
	minLevel Level
	mu       sync.Mutex
}

func New(out io.Writer, minLevel Level) *Logger {
	return &Logger{
		out:      out,
		minLevel: minLevel,
	}
}

func (l *Logger) PrintInfo(message string) {
	l.print(LevelInfo, message)
}

func (l *Logger) PrintError(err error) {
	l.print(LevelInfo, err.Error())
}

func (l *Logger) PrintFatal(err error) {
	l.print(LevelFatal, err.Error())
	os.Exit(1)
}

func (l *Logger) print(level Level, message string) (int, error) {
	if level < l.minLevel {
		return 0, nil
	}

	timestamp := time.Now().UTC().Format("2006-01-02 15:04:05")
	output := fmt.Sprintf("%s %s: %s", timestamp, level, message)

	if level >= LevelError {
		output += "\n" + string(debug.Stack())
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	return l.out.Write(append([]byte(output), '\n'))
}

func (l *Logger) Write(message []byte) (int, error) {
	return l.print(LevelError, string(message))
}
