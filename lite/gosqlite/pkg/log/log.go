package log

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// LogLevel defines the severity of a log entry.
type LogLevel int

const (
	DEBUG LogLevel = iota
	INFO
	WARN
	ERROR
	FATAL
)

// String returns the string representation of the LogLevel.
func (l LogLevel) String() string {
	switch l {
	case DEBUG:
		return "DEBUG"
	case INFO:
		return "INFO"
	case WARN:
		return "WARN"
	case ERROR:
		return "ERROR"
	case FATAL:
		return "FATAL"
	default:
		return "UNKNOWN"
	}
}

// LogEntry represents a single log entry in JSON Lines format.
type LogEntry struct {
	Timestamp string `json:"timestamp"`
	Level     string `json:"level"`
	Message   string `json:"message"`
	// Add more fields as needed, e.g., "component", "error", "fields"
}

// Logger provides logging functionality with JSON Lines format and rolling files.
type Logger struct {
	file       *os.File
	mu         sync.Mutex
	maxSize    int64 // Maximum log file size in bytes
	currentSize int64
	logDirPath string // New field to store the log directory path
}

// NewLogger creates a new Logger instance.
// logDirPath is the directory where log files will be stored.
// maxSize is the maximum size of a log file before it rolls over (in bytes).
func NewLogger(logDirPath string, maxSize int64) (*Logger, error) {
	if logDirPath == "" {
		return nil, fmt.Errorf("log output path cannot be empty")
	}
	if maxSize <= 0 {
		return nil, fmt.Errorf("max log size must be positive")
	}

	if err := os.MkdirAll(logDirPath, 0755); err != nil {
		return nil, fmt.Errorf("failed to create log directory %s: %w", logDirPath, err)
	}

	logger := &Logger{
		maxSize:    maxSize,
		logDirPath: logDirPath,
	}

	if err := logger.openLogFile(); err != nil {
		return nil, fmt.Errorf("failed to open initial log file: %w", err)
	}

	return logger, nil
}

// openLogFile opens a new log file or rolls over the existing one.
func (l *Logger) openLogFile() error {
	if l.file != nil {
		l.file.Close()
	}

	timestamp := time.Now().UTC().Format("20060102_150405")
	logFileName := fmt.Sprintf("gosqlite_%s.log", timestamp)
	logFilePath := filepath.Join(l.logDirPath, logFileName)

	file, err := os.OpenFile(logFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	l.file = file

	info, err := file.Stat()
	if err != nil {
		return err
	}
	l.currentSize = info.Size()

	return nil
}

// writeLogEntry writes a LogEntry to the current log file.
func (l *Logger) writeLogEntry(level LogLevel, message string) {
	l.mu.Lock()
	defer l.mu.Unlock()

	entry := LogEntry{
		Timestamp: time.Now().UTC().Format(time.RFC3339Nano),
		Level:     level.String(),
		Message:   message,
	}

	jsonBytes, err := json.Marshal(entry)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to marshal log entry: %v\n", err)
		return
	}

	// Check for file size and roll over if necessary
	if l.currentSize+int64(len(jsonBytes))+1 > l.maxSize { // +1 for newline
		if err := l.openLogFile(); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to roll over log file: %v\n", err)
			return
		}
	}

	if _, err := l.file.Write(jsonBytes); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to write log entry to file: %v\n", err)
		return
	}
	if _, err := l.file.WriteString("\n"); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to write newline to log file: %v\n", err)
		return
	}
	l.currentSize += int64(len(jsonBytes)) + 1
}

// Debug logs a message at DEBUG level.
func (l *Logger) Debug(format string, a ...interface{}) {
	l.writeLogEntry(DEBUG, fmt.Sprintf(format, a...))
}

// Info logs a message at INFO level.
func (l *Logger) Info(format string, a ...interface{}) {
	l.writeLogEntry(INFO, fmt.Sprintf(format, a...))
}

// Warn logs a message at WARN level.
func (l *Logger) Warn(format string, a ...interface{}) {
	l.writeLogEntry(WARN, fmt.Sprintf(format, a...))
}

// Error logs a message at ERROR level.
func (l *Logger) Error(format string, a ...interface{}) {
	l.writeLogEntry(ERROR, fmt.Sprintf(format, a...))
}

// Fatal logs a message at FATAL level and exits the application.
func (l *Logger) Fatal(format string, a ...interface{}) {
	l.writeLogEntry(FATAL, fmt.Sprintf(format, a...))
	l.Close()
	os.Exit(1)
}

// Close closes the underlying log file.
func (l *Logger) Close() error {
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.file != nil {
		return l.file.Close()
	}
	return nil
}

// Global logger instance
var defaultLogger *Logger
var once sync.Once

// Init initializes the global logger. This should be called once at application startup.
func Init(logDirPath string, maxSize int64) {
	once.Do(func() {
		var err error
		defaultLogger, err = NewLogger(logDirPath, maxSize)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to initialize logger: %v\n", err)
			// Fallback to stderr if logger initialization fails
			defaultLogger = &Logger{
				maxSize:    maxSize,
				file:       os.Stderr, // Direct to stderr as a fallback
				mu:         sync.Mutex{},
				logDirPath: logDirPath, // Store the path for potential future use, even in fallback
			}
		}
	})
}

// Debug logs a message using the global logger.
func Debug(format string, a ...interface{}) {
	if defaultLogger != nil {
		defaultLogger.Debug(format, a...)
	}
}

// Info logs a message using the global logger.
func Info(format string, a ...interface{}) {
	if defaultLogger != nil {
		defaultLogger.Info(format, a...)
	}
}

// Warn logs a message using the global logger.
func Warn(format string, a ...interface{}) {
	if defaultLogger != nil {
		defaultLogger.Warn(format, a...)
	}
}

// Error logs a message using the global logger.
func Error(format string, a ...interface{}) {
	if defaultLogger != nil {
		defaultLogger.Error(format, a...)
	}
}

// Fatal logs a message using the global logger and exits.
func Fatal(format string, a ...interface{}) {
	if defaultLogger != nil {
		defaultLogger.Fatal(format, a...)
	} else {
		fmt.Fprintf(os.Stderr, "FATAL: "+format+"\n", a...)
		os.Exit(1)
	}
}

// CloseGlobalLogger closes the global logger.
func CloseGlobalLogger() {
	if defaultLogger != nil {
		defaultLogger.Close()
	}
}
