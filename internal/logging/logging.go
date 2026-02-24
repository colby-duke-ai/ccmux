// Package logging provides simple file-based logging.
package logging

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

var logFile *os.File

func Init() error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	logDir := filepath.Join(homeDir, ".ccmux", "logs")
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return err
	}

	logPath := filepath.Join(logDir, "ccmux.log")
	logFile, err = os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return err
	}

	Log("=== ccmux started ===")
	return nil
}

func Close() {
	if logFile != nil {
		logFile.Close()
	}
}

func Log(format string, args ...interface{}) {
	if logFile == nil {
		return
	}
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	msg := fmt.Sprintf(format, args...)
	fmt.Fprintf(logFile, "[%s] %s\n", timestamp, msg)
	logFile.Sync()
}

