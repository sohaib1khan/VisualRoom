package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/spf13/viper"
)

const AppName = "vroom"

type Config struct {
	Scan          ScanConfig          `mapstructure:"scan"`
	Thresholds    Thresholds          `mapstructure:"thresholds"`
	Bender        BenderConfig        `mapstructure:"bender"`
	Cleanup       CleanupConfig       `mapstructure:"cleanup"`
	Scheduler     SchedulerConfig     `mapstructure:"scheduler"`
	Notifications NotificationsConfig `mapstructure:"notifications"`
}

type ScanConfig struct {
	MaxDepth       int      `mapstructure:"max_depth"`
	FollowSymlinks bool     `mapstructure:"follow_symlinks"`
	SkipPaths      []string `mapstructure:"skip_paths"`
}

type Thresholds struct {
	Relaxed   float64 `mapstructure:"relaxed"`
	Sarcastic float64 `mapstructure:"sarcastic"`
	Sweating  float64 `mapstructure:"sweating"`
	Panicking float64 `mapstructure:"panicking"`
}

type BenderConfig struct {
	TickMS       int  `mapstructure:"tick_ms"`
	Catchphrases bool `mapstructure:"catchphrases"`
}

type CleanupConfig struct {
	QuarantineDir string          `mapstructure:"quarantine_dir"`
	TTLDays       int             `mapstructure:"ttl_days"`
	DryRun        bool            `mapstructure:"dry_run"`
	HomeOnly      bool            `mapstructure:"home_only"`
	Protected     []string        `mapstructure:"protected"`
	AutoClean     AutoCleanConfig `mapstructure:"auto_clean"`
}

type AutoCleanConfig struct {
	MinAgeDays int      `mapstructure:"min_age_days"`
	Patterns   []string `mapstructure:"patterns"`
	FlagOnly   []string `mapstructure:"flag_only"`
}

type SchedulerConfig struct {
	Engine   string `mapstructure:"engine"`
	Interval string `mapstructure:"interval"`
	Mode     string `mapstructure:"mode"`
}

type NotificationsConfig struct {
	Desktop    bool   `mapstructure:"desktop"`
	StatusFile string `mapstructure:"status_file"`
	LoginHook  bool   `mapstructure:"login_hook"`
}

var (
	global     *Config
	globalOnce sync.Once
	globalErr  error
)

func Default() *Config {
	return &Config{
		Scan: ScanConfig{
			MaxDepth:       1,
			FollowSymlinks: false,
			SkipPaths:      []string{"/proc", "/sys", "/dev", "/run", "/snap"},
		},
		Thresholds: Thresholds{
			Relaxed:   50,
			Sarcastic: 20,
			Sweating:  10,
			Panicking: 3,
		},
		Bender: BenderConfig{
			TickMS:       110,
			Catchphrases: true,
		},
		Cleanup: CleanupConfig{
			QuarantineDir: "~/.vroom/quarantine",
			TTLDays:       7,
			DryRun:        true,
			HomeOnly:      true,
			Protected: []string{
				"~/.ssh",
				"~/.gnupg",
				"~/.vroom",
				"~/.gitconfig",
			},
			AutoClean: AutoCleanConfig{
				MinAgeDays: 30,
				Patterns:   []string{"*.tmp", "*.log", "*.swp"},
				FlagOnly:   []string{"node_modules", ".cache", "__pycache__", ".tox", ".venv"},
			},
		},
		Scheduler: SchedulerConfig{
			Engine:   "native",
			Interval: "daily",
			Mode:     "notify",
		},
		Notifications: NotificationsConfig{
			Desktop:    true,
			StatusFile: "~/.vroom/status.txt",
			LoginHook:  false,
		},
	}
}

func DataDir() string {
	if home, err := os.UserHomeDir(); err == nil && home != "" {
		return filepath.Join(home, ".vroom")
	}
	return filepath.Join(os.TempDir(), ".vroom")
}

func HomeDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return home
}

func ExpandPath(p string) string {
	if p == "" {
		return p
	}
	if p == "~" {
		return HomeDir()
	}
	if strings.HasPrefix(p, "~/") {
		return filepath.Join(HomeDir(), p[2:])
	}
	if strings.HasPrefix(p, "$HOME/") || strings.HasPrefix(p, "${HOME}/") {
		p = strings.Replace(p, "${HOME}", HomeDir(), 1)
		p = strings.Replace(p, "$HOME", HomeDir(), 1)
	}
	return p
}

func EnsureDataDir() error {
	dir := DataDir()
	return os.MkdirAll(dir, 0o700)
}

func UserConfigPath() string {
	return filepath.Join(DataDir(), "config.yaml")
}

func Load() (*Config, error) {
	globalOnce.Do(func() {
		global, globalErr = load()
	})
	return global, globalErr
}

func MustLoad() *Config {
	cfg, err := Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "vroom: config warning: %v (using defaults)\n", err)
		return Default()
	}
	return cfg
}

func load() (*Config, error) {
	cfg := Default()
	_ = EnsureDataDir()

	v := viper.New()
	v.SetConfigType("yaml")
	v.SetEnvPrefix("VROOM")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	userPath := UserConfigPath()
	if _, err := os.Stat(userPath); os.IsNotExist(err) {
		if seed := findSeedYAML(); seed != "" {
			_ = copyFile(seed, userPath)
		}
	}

	v.SetConfigName("config")
	v.AddConfigPath(DataDir())
	v.AddConfigPath("./config")
	v.AddConfigPath(".")

	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return cfg, err
		}
	}

	if err := v.Unmarshal(cfg); err != nil {
		return cfg, err
	}
	cfg.Cleanup.QuarantineDir = ExpandPath(cfg.Cleanup.QuarantineDir)
	cfg.Notifications.StatusFile = ExpandPath(cfg.Notifications.StatusFile)
	for i, p := range cfg.Cleanup.Protected {
		cfg.Cleanup.Protected[i] = ExpandPath(p)
	}
	for i, p := range cfg.Scan.SkipPaths {
		cfg.Scan.SkipPaths[i] = ExpandPath(p)
	}
	return cfg, nil
}

func findSeedYAML() string {
	candidates := []string{
		"config/default.yaml",
		filepath.Join("config", "default.yaml"),
	}
	if exe, err := os.Executable(); err == nil {
		candidates = append(candidates, filepath.Join(filepath.Dir(exe), "config", "default.yaml"))
	}
	for _, c := range candidates {
		if st, err := os.Stat(c); err == nil && !st.IsDir() {
			return c
		}
	}
	return ""
}

func copyFile(src, dst string) error {
	data, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	return os.WriteFile(dst, data, 0o600)
}
