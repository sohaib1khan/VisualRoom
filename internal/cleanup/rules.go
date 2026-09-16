package cleanup

import (
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"strings"

	"vroom/internal/config"
	"vroom/internal/scanner"
)

type Decision struct {
	Path    string
	Allowed bool
	Reason  string
}

func HomeDir() string {
	if h, err := os.UserHomeDir(); err == nil {
		return h
	}
	u, err := user.Current()
	if err != nil {
		return ""
	}
	return u.HomeDir
}

func IsProtected(path string, cfg *config.Config) Decision {
	abs, err := filepath.Abs(path)
	if err != nil {
		return Decision{Path: path, Allowed: false, Reason: "cannot resolve path"}
	}
	abs = filepath.Clean(abs)

	if abs == "/" {
		return Decision{Path: abs, Allowed: false, Reason: "refusing to touch filesystem root"}
	}

	home := HomeDir()
	if cfg.Cleanup.HomeOnly && home != "" {
		homeAbs := filepath.Clean(home)
		if abs != homeAbs && !strings.HasPrefix(abs, homeAbs+string(os.PathSeparator)) {
			return Decision{Path: abs, Allowed: false, Reason: "path is outside $HOME (home_only is on)"}
		}
	}

	system := []string{
		"/bin", "/sbin", "/usr", "/etc", "/boot", "/lib", "/lib64",
		"/dev", "/proc", "/sys", "/run", "/root",
	}
	for _, p := range system {
		if abs == p || strings.HasPrefix(abs, p+string(os.PathSeparator)) {
			return Decision{Path: abs, Allowed: false, Reason: "system path is protected"}
		}
	}

	for _, p := range cfg.Cleanup.Protected {
		prot := filepath.Clean(config.ExpandPath(p))
		if prot == "" {
			continue
		}
		if abs == prot || strings.HasPrefix(abs, prot+string(os.PathSeparator)) {
			return Decision{Path: abs, Allowed: false, Reason: "matches protected path " + prot}
		}
	}

	base := filepath.Base(abs)
	if base == ".git" || base == ".ssh" || base == ".gnupg" {
		return Decision{Path: abs, Allowed: false, Reason: "refusing to touch " + base}
	}
	if strings.Contains(abs, string(os.PathSeparator)+".git"+string(os.PathSeparator)) {
		return Decision{Path: abs, Allowed: false, Reason: "refusing to touch files inside .git"}
	}

	qdir := filepath.Clean(cfg.Cleanup.QuarantineDir)
	if qdir != "" && (abs == qdir || strings.HasPrefix(abs, qdir+string(os.PathSeparator))) {
		return Decision{Path: abs, Allowed: false, Reason: "cannot quarantine the quarantine directory"}
	}

	data := filepath.Clean(config.DataDir())
	if abs == data {
		return Decision{Path: abs, Allowed: false, Reason: "cannot remove vroom data directory"}
	}

	return Decision{Path: abs, Allowed: true}
}

func RequireElevatedOK(allowElevated bool) error {
	if os.Geteuid() == 0 && !allowElevated {
		return fmt.Errorf("refusing to run cleanup as root without --allow-elevated (vroom never auto-escalates)")
	}
	return nil
}

func DockerCleanBlocked(path string, force bool) error {
	if !scanner.InContainer() {
		return nil
	}
	if scanner.IsReadOnlyMount(path) {
		return fmt.Errorf("docker mode: %s is a read-only mount; cleanup blocked (mount :rw and pass --force)", path)
	}
	if !force {
		return fmt.Errorf("docker mode: cleanup requires a writable mount and --force")
	}
	return nil
}

func MatchesAutoClean(path string, isDir bool, cfg *config.Config) bool {
	if isDir {
		return false
	}
	name := filepath.Base(path)
	for _, pat := range cfg.Cleanup.AutoClean.Patterns {
		ok, err := filepath.Match(pat, name)
		if err == nil && ok {
			return true
		}
	}
	return false
}

func IsFlagOnly(path string, cfg *config.Config) bool {
	base := filepath.Base(path)
	for _, name := range cfg.Cleanup.AutoClean.FlagOnly {
		if base == name {
			return true
		}
	}
	return false
}
