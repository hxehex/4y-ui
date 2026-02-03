// Package config provides configuration management utilities for the 4y-ui panel,
// including version information, logging levels, database paths, and environment variable handling.
package config

import (
	_ "embed"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

//go:embed version
var version string

//go:embed name
var name string

// LogLevel represents the logging level for the application.
type LogLevel string

// Logging level constants
const (
	Debug   LogLevel = "debug"
	Info    LogLevel = "info"
	Notice  LogLevel = "notice"
	Warning LogLevel = "warning"
	Error   LogLevel = "error"
)

// GetVersion returns the version string of the 4y-ui application.
func GetVersion() string {
	return strings.TrimSpace(version)
}

// GetName returns the name of the 4y-ui application.
func GetName() string {
	return strings.TrimSpace(name)
}

// GetLogLevel returns the current logging level based on environment variables or defaults to Info.
func GetLogLevel() LogLevel {
	if IsDebug() {
		return Debug
	}
	logLevel := os.Getenv("XUI_LOG_LEVEL")
	if logLevel == "" {
		return Info
	}
	return LogLevel(logLevel)
}

// IsDebug returns true if debug mode is enabled via the XUI_DEBUG environment variable.
func IsDebug() bool {
	return os.Getenv("XUI_DEBUG") == "true"
}

// GetBinFolderPath returns the path to the binary folder, defaulting to "bin" if not set via XUI_BIN_FOLDER.
func GetBinFolderPath() string {
	binFolderPath := os.Getenv("XUI_BIN_FOLDER")
	if binFolderPath == "" {
		binFolderPath = "bin"
	}
	return binFolderPath
}

func getBaseDir() string {
	exePath, err := os.Executable()
	if err != nil {
		return "."
	}
	exeDir := filepath.Dir(exePath)
	exeDirLower := strings.ToLower(filepath.ToSlash(exeDir))
	if strings.Contains(exeDirLower, "/appdata/local/temp/") || strings.Contains(exeDirLower, "/go-build") {
		wd, err := os.Getwd()
		if err != nil {
			return "."
		}
		return wd
	}
	return exeDir
}

// GetDBFolderPath returns the path to the database folder.
// It prioritizes Env Var -> Windows Base Dir -> XDG Config Home (~/.config/4y-ui).
func GetDBFolderPath() string {
	// 1. Environment Variable
	dbFolderPath := os.Getenv("XUI_DB_FOLDER")
	if dbFolderPath != "" {
		return dbFolderPath
	}

	// 2. Windows specific logic
	if runtime.GOOS == "windows" {
		return getBaseDir()
	}

	// 3. XDG Standard (Linux/macOS)
	// Try to use ~/.config/4y-ui
	configDir, err := os.UserConfigDir()
	if err != nil {
		// Fallback to local directory if user config dir is unavailable
		return "./"
	}

	targetDir := filepath.Join(configDir, "4y-ui")

	// Ensure directory exists so we don't get write errors later
	_ = os.MkdirAll(targetDir, 0755)

	return targetDir
}

// GetDBPath returns the full path to the database file.
func GetDBPath() string {
	return fmt.Sprintf("%s/%s.db", GetDBFolderPath(), GetName())
}

// GetLogFolder returns the path to the log folder.
// It prioritizes Env Var -> Windows Local -> XDG State Home (~/.local/state/4y-ui).
func GetLogFolder() string {
	// 1. Environment Variable
	logFolderPath := os.Getenv("XUI_LOG_FOLDER")
	if logFolderPath != "" {
		return logFolderPath
	}

	// 2. Windows specific logic
	if runtime.GOOS == "windows" {
		return filepath.Join(".", "log")
	}

	// 3. XDG Standard (Linux/macOS)
	// Modern Linux standard for logs is ~/.local/state or ~/.cache
	homeDir, err := os.UserHomeDir()
	if err != nil {
		// Fallback to local
		return "./"
	}

	// Construct ~/.local/state/4y-ui
	targetLogDir := filepath.Join(homeDir, ".local", "state", "4y-ui")

	// Ensure directory exists
	_ = os.MkdirAll(targetLogDir, 0755)

	return targetLogDir
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, in)
	if err != nil {
		return err
	}

	return out.Sync()
}

func init() {
	// Migration logic specifically for Windows or custom needs.
	// We leave this alone to avoid breaking Windows backward compatibility logic from original repo.
	if runtime.GOOS != "windows" {
		return
	}
	if os.Getenv("XUI_DB_FOLDER") != "" {
		return
	}
	oldDBFolder := "/etc/4y-ui"
	oldDBPath := fmt.Sprintf("%s/%s.db", oldDBFolder, GetName())
	newDBFolder := GetDBFolderPath()
	newDBPath := fmt.Sprintf("%s/%s.db", newDBFolder, GetName())
	_, err := os.Stat(newDBPath)
	if err == nil {
		return // new exists
	}
	_, err = os.Stat(oldDBPath)
	if os.IsNotExist(err) {
		return // old does not exist
	}
	_ = copyFile(oldDBPath, newDBPath) // ignore error
}
