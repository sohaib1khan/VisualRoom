package cleanup

import (
	"os"
	"path/filepath"
	"testing"

	"vroom/internal/config"
)

func TestIsProtectedHomeOnly(t *testing.T) {
	cfg := config.Default()
	cfg.Cleanup.HomeOnly = true
	home, err := os.UserHomeDir()
	if err != nil {
		t.Skip(err)
	}
	cfg.Cleanup.Protected = []string{filepath.Join(home, ".ssh")}
	cfg.Cleanup.QuarantineDir = filepath.Join(home, ".vroom", "quarantine")

	if d := IsProtected("/etc/passwd", cfg); d.Allowed {
		t.Fatal("expected /etc/passwd blocked")
	}
	if d := IsProtected(filepath.Join(home, ".ssh", "id_rsa"), cfg); d.Allowed {
		t.Fatal("expected ~/.ssh blocked")
	}
	target := filepath.Join(home, "Downloads", "old.zip")
	if d := IsProtected(target, cfg); !d.Allowed {
		t.Fatalf("expected downloads allowed, got %s", d.Reason)
	}
}

func TestIsProtectedRoot(t *testing.T) {
	cfg := config.Default()
	if d := IsProtected("/", cfg); d.Allowed {
		t.Fatal("root must be blocked")
	}
}
