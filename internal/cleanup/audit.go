package cleanup

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"vroom/internal/config"
)

var auditMu sync.Mutex

func AuditPath() string {
	return filepath.Join(config.DataDir(), "audit.log")
}

func AppendAudit(line string) error {
	if err := config.EnsureDataDir(); err != nil {
		return err
	}
	auditMu.Lock()
	defer auditMu.Unlock()
	f, err := os.OpenFile(AuditPath(), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o600)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = fmt.Fprintf(f, "%s %s\n", time.Now().UTC().Format(time.RFC3339), line)
	return err
}
