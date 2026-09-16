//go:build linux

package scanner

import (
	"os"
	"time"

	"golang.org/x/sys/unix"
)

func extraMeta(path string, info os.FileInfo) (atime time.Time, owner string) {
	atime = info.ModTime()
	owner = "?"
	var st unix.Stat_t
	if err := unix.Lstat(path, &st); err != nil {
		return atime, owner
	}
	atime = time.Unix(int64(st.Atim.Sec), int64(st.Atim.Nsec))
	owner = lookupOwner(st.Uid)
	return atime, owner
}

func statfs(path string) (DiskUsage, error) {
	var st unix.Statfs_t
	if err := unix.Statfs(path, &st); err != nil {
		return DiskUsage{}, err
	}
	bsize := uint64(st.Bsize)
	total := st.Blocks * bsize
	free := st.Bfree * bsize
	avail := st.Bavail * bsize
	used := uint64(0)
	if total > free {
		used = total - free
	}
	return DiskUsage{
		Path:  path,
		Total: total,
		Free:  free,
		Avail: avail,
		Used:  used,
	}, nil
}

func IsReadOnlyMount(path string) bool {
	var st unix.Statfs_t
	if err := unix.Statfs(path, &st); err != nil {
		return false
	}
	return st.Flags&unix.ST_RDONLY != 0
}
