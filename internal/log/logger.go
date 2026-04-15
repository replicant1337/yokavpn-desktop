package log

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

type Level int

const (
	DEBUG Level = iota
	INFO
	WARN
	ERROR
)

var levelNames = []string{"DEBUG", "INFO", "WARN", "ERROR"}

type Logger struct {
	mu       sync.Mutex
	file     *os.File
	module   string
	minLevel Level
	logChan  chan string
	console  bool
}

var loggers = make(map[string]*Logger)
var defaultLogger *Logger

func init() {
	dir := filepath.Join(os.Getenv("APPDATA"), "YokoVPN", "logs")
	os.MkdirAll(dir, 0755)

	var err error
	defaultLogger, err = NewLogger("YokoVPN", dir)
	if err != nil {
		fmt.Printf("Failed to create logger: %v\n", err)
	}
}

func NewLogger(module string, logDir string) (*Logger, error) {
	if _, ok := loggers[module]; ok {
		return loggers[module], nil
	}

	filename := fmt.Sprintf("%s_%s.log", module, time.Now().Format("2006-01-02"))
	path := filepath.Join(logDir, filename)

	file, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return nil, err
	}

	logger := &Logger{
		file:     file,
		module:   module,
		minLevel: DEBUG,
		logChan:  make(chan string, 1000),
		console:  true,
	}

	loggers[module] = logger
	go logger.writeLoop()

	return logger, nil
}

func GetLogger(module string) *Logger {
	if l, ok := loggers[module]; ok {
		return l
	}
	return defaultLogger
}

func (l *Logger) SetLevel(level Level) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.minLevel = level
}

func (l *Logger) SetConsole(enabled bool) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.console = enabled
}

func (l *Logger) writeLoop() {
	for msg := range l.logChan {
		l.mu.Lock()
		if l.file != nil {
			l.file.WriteString(msg + "\n")
		}
		l.mu.Unlock()
	}
}

func (l *Logger) log(level Level, format string, args ...interface{}) {
	if level < l.minLevel {
		return
	}

	timestamp := time.Now().Format("2006-01-02 15:04:05.000")
	msg := fmt.Sprintf(format, args...)
	logLine := fmt.Sprintf("[%s] [%s] [%s] %s", timestamp, levelNames[level], l.module, msg)

	l.logChan <- logLine

	if l.console {
		fmt.Println(logLine)
	}
}

func (l *Logger) Debug(format string, args ...interface{}) {
	l.log(DEBUG, format, args...)
}

func (l *Logger) Info(format string, args ...interface{}) {
	l.log(INFO, format, args...)
}

func (l *Logger) Warn(format string, args ...interface{}) {
	l.log(WARN, format, args...)
}

func (l *Logger) Error(format string, args ...interface{}) {
	l.log(ERROR, format, args...)
}

func (l *Logger) Close() {
	l.mu.Lock()
	defer l.mu.Unlock()
	close(l.logChan)
	if l.file != nil {
		l.file.Close()
	}
}

func Debug(format string, args ...interface{}) {
	defaultLogger.Debug(format, args...)
}

func Info(format string, args ...interface{}) {
	defaultLogger.Info(format, args...)
}

func Warn(format string, args ...interface{}) {
	defaultLogger.Warn(format, args...)
}

func Error(format string, args ...interface{}) {
	defaultLogger.Error(format, args...)
}

func GetLogDir() string {
	return filepath.Join(os.Getenv("APPDATA"), "YokoVPN", "logs")
}
