package cleanup

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"vroom/internal/config"
)

type Item struct {
	ID         string `json:"id"`
	Original   string `json:"original"`
	Quarantine string `json:"quarantine"`
	Size       int64  `json:"size"`
	IsDir      bool   `json:"is_dir"`
}

type Batch struct {
	ID        string    `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	ExpiresAt time.Time `json:"expires_at"`
	Items     []Item    `json:"items"`
	DryRun    bool      `json:"dry_run"`
}

type Result struct {
	Batch   *Batch
	Skipped []Decision
	DryRun  bool
	Moved   int
	Bytes   int64
}

func newID() string {
	return time.Now().UTC().Format("20060102-150405")
}

func batchDir(cfg *config.Config, id string) string {
	return filepath.Join(cfg.Cleanup.QuarantineDir, id)
}

func LoadBatch(cfg *config.Config, id string) (*Batch, error) {
	p := filepath.Join(batchDir(cfg, id), "manifest.json")
	data, err := os.ReadFile(p)
	if err != nil {
		return nil, err
	}
	var b Batch
	if err := json.Unmarshal(data, &b); err != nil {
		return nil, err
	}
	return &b, nil
}

func ListBatches(cfg *config.Config) ([]Batch, error) {
	root := cfg.Cleanup.QuarantineDir
	entries, err := os.ReadDir(root)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	var out []Batch
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		b, err := LoadBatch(cfg, e.Name())
		if err != nil {
			continue
		}
		out = append(out, *b)
	}
	return out, nil
}

func Quarantine(cfg *config.Config, paths []string, dryRun, allowElevated bool) (*Result, error) {
	if !dryRun {
		if err := RequireElevatedOK(allowElevated); err != nil {
			return nil, err
		}
	}
	id := newID()
	ttl := time.Duration(cfg.Cleanup.TTLDays) * 24 * time.Hour
	if ttl <= 0 {
		ttl = 7 * 24 * time.Hour
	}
	batch := &Batch{
		ID:        id,
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().Add(ttl),
		DryRun:    dryRun,
	}
	res := &Result{Batch: batch, DryRun: dryRun}

	destRoot := batchDir(cfg, id)
	if !dryRun {
		if err := os.MkdirAll(filepath.Join(destRoot, "files"), 0o700); err != nil {
			return nil, err
		}
	}

	for i, raw := range paths {
		abs, err := filepath.Abs(raw)
		if err != nil {
			res.Skipped = append(res.Skipped, Decision{Path: raw, Allowed: false, Reason: err.Error()})
			continue
		}
		d := IsProtected(abs, cfg)
		if !d.Allowed {
			res.Skipped = append(res.Skipped, d)
			continue
		}
		info, err := os.Lstat(abs)
		if err != nil {
			res.Skipped = append(res.Skipped, Decision{Path: abs, Allowed: false, Reason: err.Error()})
			continue
		}
		item := Item{
			ID:       fmt.Sprintf("%s-%02d", id, i+1),
			Original: abs,
			Size:     sizeOf(abs, info),
			IsDir:    info.IsDir(),
		}
		rel := strings.TrimPrefix(abs, string(os.PathSeparator))
		qpath := filepath.Join(destRoot, "files", rel)
		item.Quarantine = qpath
		if dryRun {
			batch.Items = append(batch.Items, item)
			res.Moved++
			res.Bytes += item.Size
			continue
		}
		if err := os.MkdirAll(filepath.Dir(qpath), 0o700); err != nil {
			res.Skipped = append(res.Skipped, Decision{Path: abs, Allowed: false, Reason: err.Error()})
			continue
		}
		if err := movePath(abs, qpath); err != nil {
			res.Skipped = append(res.Skipped, Decision{Path: abs, Allowed: false, Reason: err.Error()})
			continue
		}
		batch.Items = append(batch.Items, item)
		res.Moved++
		res.Bytes += item.Size
		_ = AppendAudit(fmt.Sprintf("QUARANTINE id=%s path=%s size=%d dest=%s", batch.ID, abs, item.Size, qpath))
	}

	if !dryRun && len(batch.Items) > 0 {
		data, err := json.MarshalIndent(batch, "", "  ")
		if err != nil {
			return res, err
		}
		if err := os.WriteFile(filepath.Join(destRoot, "manifest.json"), data, 0o600); err != nil {
			return res, err
		}
	} else if dryRun {
		_ = AppendAudit(fmt.Sprintf("QUARANTINE-DRY-RUN id=%s items=%d bytes=%d", batch.ID, res.Moved, res.Bytes))
	}
	return res, nil
}

func Restore(cfg *config.Config, id string, allowElevated bool) error {
	if err := RequireElevatedOK(allowElevated); err != nil {
		return err
	}
	batch, err := LoadBatch(cfg, id)
	if err != nil {
		return fmt.Errorf("unknown quarantine id %q: %w", id, err)
	}
	var first error
	for _, item := range batch.Items {
		if _, err := os.Stat(item.Original); err == nil {
			if first == nil {
				first = fmt.Errorf("restore blocked: %s already exists", item.Original)
			}
			continue
		}
		if err := os.MkdirAll(filepath.Dir(item.Original), 0o755); err != nil {
			if first == nil {
				first = err
			}
			continue
		}
		if err := movePath(item.Quarantine, item.Original); err != nil {
			if first == nil {
				first = err
			}
			continue
		}
		_ = AppendAudit(fmt.Sprintf("RESTORE id=%s path=%s from=%s", id, item.Original, item.Quarantine))
	}
	_ = os.RemoveAll(batchDir(cfg, id))
	return first
}

func PurgeExpired(cfg *config.Config, force bool) (int, error) {
	batches, err := ListBatches(cfg)
	if err != nil {
		return 0, err
	}
	n := 0
	now := time.Now()
	for _, b := range batches {
		if !force && now.Before(b.ExpiresAt) {
			continue
		}
		if err := os.RemoveAll(batchDir(cfg, b.ID)); err != nil {
			return n, err
		}
		_ = AppendAudit(fmt.Sprintf("PURGE id=%s items=%d", b.ID, len(b.Items)))
		n++
	}
	return n, nil
}

func movePath(src, dst string) error {
	if err := os.Rename(src, dst); err == nil {
		return nil
	}
	if err := copyTree(src, dst); err != nil {
		return err
	}
	return os.RemoveAll(src)
}

func copyTree(src, dst string) error {
	info, err := os.Lstat(src)
	if err != nil {
		return err
	}
	if info.Mode()&os.ModeSymlink != 0 {
		target, err := os.Readlink(src)
		if err != nil {
			return err
		}
		return os.Symlink(target, dst)
	}
	if info.IsDir() {
		if err := os.MkdirAll(dst, info.Mode()); err != nil {
			return err
		}
		entries, err := os.ReadDir(src)
		if err != nil {
			return err
		}
		for _, e := range entries {
			if err := copyTree(filepath.Join(src, e.Name()), filepath.Join(dst, e.Name())); err != nil {
				return err
			}
		}
		return nil
	}
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return err
	}
	out, err := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, info.Mode())
	if err != nil {
		return err
	}
	defer out.Close()
	_, err = io.Copy(out, in)
	return err
}

func sizeOf(path string, info os.FileInfo) int64 {
	if !info.IsDir() {
		return info.Size()
	}
	var total int64
	_ = filepath.Walk(path, func(_ string, fi os.FileInfo, err error) error {
		if err != nil || fi == nil || fi.IsDir() {
			return nil
		}
		total += fi.Size()
		return nil
	})
	return total
}
