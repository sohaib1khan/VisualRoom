package scanner

import (
	"context"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"sync"
	"syscall"
	"time"

	"vroom/internal/config"
)

type Node struct {
	Path     string
	Name     string
	Size     int64
	IsDir    bool
	Count    int
	Mode     os.FileMode
	ModTime  time.Time
	Atime    time.Time
	Owner    string
	Children []*Node
	Err      error
}

type DiskUsage struct {
	Path  string
	Total uint64
	Free  uint64
	Avail uint64
	Used  uint64
}

func (d DiskUsage) FreePct() float64 {
	if d.Total == 0 {
		return 0
	}
	return 100 * float64(d.Avail) / float64(d.Total)
}

func (d DiskUsage) UsedPct() float64 {
	if d.Total == 0 {
		return 0
	}
	return 100 * float64(d.Used) / float64(d.Total)
}

type Options struct {
	FollowSymlinks bool
	SkipPaths      []string
	MaxDepth       int
	Workers        int
	Cache          *Cache
}

func DefaultOptions() Options {
	cfg := config.MustLoad()
	return Options{
		FollowSymlinks: cfg.Scan.FollowSymlinks,
		SkipPaths:      cfg.Scan.SkipPaths,
		MaxDepth:       cfg.Scan.MaxDepth,
		Workers:        runtime.NumCPU() * 2,
		Cache:          DefaultCache(),
	}
}

type Walker struct {
	opts Options
}

func New(opts Options) *Walker {
	if opts.Workers < 1 {
		opts.Workers = runtime.NumCPU() * 2
	}
	return &Walker{opts: opts}
}

func Default() *Walker {
	return New(DefaultOptions())
}

func (w *Walker) shouldSkip(path string) bool {
	clean := filepath.Clean(path)
	for _, s := range w.opts.SkipPaths {
		if s == "" {
			continue
		}
		skip := filepath.Clean(s)
		if clean == skip || strings.HasPrefix(clean, skip+string(os.PathSeparator)) {
			return true
		}
	}
	return false
}

// ScanDir lists immediate children of path and computes each child's recursive size.
func (w *Walker) ScanDir(ctx context.Context, path string) (*Node, error) {
	abs, err := filepath.Abs(path)
	if err != nil {
		return nil, err
	}
	info, err := os.Lstat(abs)
	if err != nil {
		return nil, err
	}
	root := nodeFromInfo(abs, info)
	if !info.IsDir() {
		root.Count = 1
		return root, nil
	}

	entries, err := os.ReadDir(abs)
	if err != nil {
		root.Err = err
		return root, err
	}

	type result struct {
		node *Node
	}
	workers := w.opts.Workers
	if workers > len(entries) && len(entries) > 0 {
		workers = len(entries)
	}
	sem := make(chan struct{}, workers)
	var wg sync.WaitGroup
	out := make(chan result, len(entries))

	for _, e := range entries {
		e := e
		childPath := filepath.Join(abs, e.Name())
		if w.shouldSkip(childPath) {
			continue
		}
		if ctx.Err() != nil {
			break
		}
		wg.Add(1)
		go func() {
			defer wg.Done()
			select {
			case sem <- struct{}{}:
			case <-ctx.Done():
				return
			}
			defer func() { <-sem }()
			n := w.scanChild(ctx, childPath, e)
			out <- result{node: n}
		}()
	}

	go func() {
		wg.Wait()
		close(out)
	}()

	var children []*Node
	var total int64
	var count int
	for r := range out {
		if r.node == nil {
			continue
		}
		children = append(children, r.node)
		total += r.node.Size
		count += r.node.Count
	}
	sort.Slice(children, func(i, j int) bool {
		if children[i].Size == children[j].Size {
			return children[i].Name < children[j].Name
		}
		return children[i].Size > children[j].Size
	})
	root.Children = children
	root.Size = total
	root.Count = count
	if w.opts.Cache != nil {
		_ = w.opts.Cache.Save()
	}
	return root, nil
}

func (w *Walker) SaveCache() error {
	if w.opts.Cache == nil {
		return nil
	}
	return w.opts.Cache.Save()
}

func (w *Walker) scanChild(ctx context.Context, path string, e os.DirEntry) *Node {
	info, err := e.Info()
	if err != nil {
		return &Node{Path: path, Name: e.Name(), Err: err}
	}
	n := nodeFromInfo(path, info)
	if !info.IsDir() {
		n.Count = 1
		return n
	}
	if isSymlink(info) && !w.opts.FollowSymlinks {
		n.Count = 1
		return n
	}
	if w.opts.Cache != nil {
		if cached, ok := w.opts.Cache.Lookup(path, info.ModTime(), info.Size()); ok {
			n.Size = cached.Size
			n.Count = cached.Count
			return n
		}
	}
	size, count, walkErr := w.walkSize(ctx, path)
	n.Size = size
	n.Count = count
	n.Err = walkErr
	if w.opts.Cache != nil && walkErr == nil {
		w.opts.Cache.Store(path, info.ModTime(), info.Size(), size, count)
	}
	return n
}

func (w *Walker) walkSize(ctx context.Context, root string) (int64, int, error) {
	var total int64
	var count int
	var firstErr error
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			if errors.Is(err, os.ErrPermission) || errors.Is(err, syscall.EACCES) {
				if firstErr == nil {
					firstErr = err
				}
				if d != nil && d.IsDir() {
					return fs.SkipDir
				}
				return nil
			}
			if errors.Is(err, os.ErrNotExist) {
				return nil
			}
			return err
		}
		if ctx.Err() != nil {
			return ctx.Err()
		}
		if path != root && w.shouldSkip(path) {
			if d.IsDir() {
				return fs.SkipDir
			}
			return nil
		}
		if d.Type()&os.ModeSymlink != 0 && !w.opts.FollowSymlinks {
			info, lerr := d.Info()
			if lerr == nil && !info.IsDir() {
				total += info.Size()
				count++
			}
			if d.IsDir() || d.Type()&os.ModeDir != 0 {
				return fs.SkipDir
			}
			return nil
		}
		info, ierr := d.Info()
		if ierr != nil {
			return nil
		}
		if !info.IsDir() {
			total += info.Size()
			count++
		} else {
			count++
		}
		return nil
	})
	if err != nil && firstErr == nil {
		firstErr = err
	}
	return total, count, firstErr
}

func nodeFromInfo(path string, info os.FileInfo) *Node {
	n := &Node{
		Path:    path,
		Name:    filepath.Base(path),
		Size:    info.Size(),
		IsDir:   info.IsDir(),
		Mode:    info.Mode(),
		ModTime: info.ModTime(),
	}
	n.Atime, n.Owner = extraMeta(path, info)
	return n
}

func isSymlink(info os.FileInfo) bool {
	return info.Mode()&os.ModeSymlink != 0
}

func Usage(path string) (DiskUsage, error) {
	abs, err := filepath.Abs(path)
	if err != nil {
		return DiskUsage{}, err
	}
	return statfs(abs)
}
