//go:build !linux

package scanner

import (
	"os"
	"syscall"
	"time"
)

func extraMeta(path string, info os.FileInfo) (atime time.Time, owner string) {
	atime = info.ModTime()
	if st, ok := info.Sys().(*syscall.Stat_t); ok {
		atime = time.Unix(st.Atim.Unix())
		owner = lookupOwner(st.Uid)
		return atime, owner
	}
	return atime, "?"
}

func statfs(path string) (DiskUsage, error) {
	var st syscall.Statfs_t
	if err := syscall.Statfs(path, &st); err != nil {
		return DiskUsage{}, err
	}
	bsize := uint64(st.Bsize)
	total := uint64(st.Blocks) * bsize
	free := uint64(st.Bfree) * bsize
	avail := uint64(st.Bavail) * bsize
	used := uint64(0)
	if total > free {
		used = total - free
	}
	return DiskUsage{Path: path, Total: total, Free: free, Avail: avail, Used: used}, nil
}

func IsReadOnlyMount(path string) bool {
	return false
}
