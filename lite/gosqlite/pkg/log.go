package pkg

// Minimal logger that satisfies Phase-1 requirement for UTC, append-only
// structured log files with naïve size-based rotation.  The implementation is
// intentionally dependency-free (std-lib only) and safe for concurrent use.  It
// is *not* meant to be a fully-featured zap/logrus replacement – higher
// fidelity logging will arrive in later phases.

import (
    "encoding/json"
    "fmt"
    "os"
    "path/filepath"
    "sync"
    "time"
)

const (
    maxLogBytes   = 10 * 1024 * 1024 // 10 MiB
    maxLogBackups = 14              // keep 14 rotated files
)

// logger is a singleton to keep API surface minimal.
var (
    logInit sync.Once
    lg      *rotatingLogger
)

// rotatingLogger writes JSON events to <root>/gosqlite.log and rotates when the
// active file exceeds maxLogBytes.
type rotatingLogger struct {
    mu       sync.Mutex // guards everything below
    file     *os.File
    rootPath string
}

// initLogger initialises the global logger.  The path is resolved lazily at
// first use so callers do not need to worry about ordering.
func initLogger() {
    logRoot := defaultLogDir()
    if err := os.MkdirAll(logRoot, 0o755); err != nil {
        // fall back to C:\Temp or /tmp if ProgramData/var/log is not writable
        logRoot = os.TempDir()
    }
    f, err := os.OpenFile(filepath.Join(logRoot, "gosqlite.log"), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
    if err != nil {
        // final fallback: stderr logger – we cannot fail completely because the
        // driver must continue to work.
        lg = &rotatingLogger{file: os.Stderr, rootPath: logRoot}
        return
    }
    lg = &rotatingLogger{file: f, rootPath: logRoot}
}

// defaultLogDir returns %ProgramData%\gosqlite\logs on Windows and
// /var/log/gosqlite on other platforms.
func defaultLogDir() string {
    if dir := os.Getenv("GOSQLITE_LOG_DIR"); dir != "" {
        return dir
    }
    if os.PathSeparator == '\\' { // windows
        programData := os.Getenv("ProgramData")
        if programData == "" {
            programData = `C:\ProgramData`
        }
        return filepath.Join(programData, "gosqlite", "logs")
    }
    return "/var/log/gosqlite"
}

// Log writes a structured log record.  It is safe for concurrent use.  The
// record map MUST be JSON-marshallable.
func Log(level, component, msg string, kv map[string]interface{}) {
    logInit.Do(initLogger)
    event := map[string]interface{}{
        "ts":        time.Now().UTC().Format(time.RFC3339Nano),
        "level":     level,
        "component": component,
        "msg":       msg,
    }
    for k, v := range kv {
        event[k] = v
    }

    b, err := json.Marshal(event)
    if err != nil {
        fmt.Fprintf(os.Stderr, "gosqlite: failed to marshal log: %v\n", err)
        return
    }
    lg.writeLine(string(b))
}

// writeLine appends the given line plus newline to the active file, rotating if
// needed.
func (l *rotatingLogger) writeLine(line string) {
    l.mu.Lock()
    defer l.mu.Unlock()

    // rotate if size would exceed threshold
    if fi, err := l.file.Stat(); err == nil && fi.Size()+int64(len(line)+1) > maxLogBytes {
        _ = l.file.Close()
        for i := maxLogBackups - 1; i >= 1; i-- {
            old := filepath.Join(l.rootPath, fmt.Sprintf("gosqlite.log.%02d", i))
            newer := filepath.Join(l.rootPath, fmt.Sprintf("gosqlite.log.%02d", i+1))
            if _, err := os.Stat(old); err == nil {
                _ = os.Rename(old, newer)
            }
        }
        // move current to .01
        _ = os.Rename(filepath.Join(l.rootPath, "gosqlite.log"), filepath.Join(l.rootPath, "gosqlite.log.01"))

        f, err := os.OpenFile(filepath.Join(l.rootPath, "gosqlite.log"), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
        if err == nil {
            l.file = f
        } else {
            // fallback to stderr to avoid panic
            l.file = os.Stderr
        }
    }

    fmt.Fprintln(l.file, line)
}
