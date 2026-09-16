package scanner

import (
	"os"
	"os/user"
	"strconv"
)

func lookupOwner(uid uint32) string {
	id := strconv.FormatUint(uint64(uid), 10)
	u, err := user.LookupId(id)
	if err != nil {
		return id
	}
	if u.Username != "" {
		return u.Username
	}
	return id
}

func InContainer() bool {
	if _, err := os.Stat("/.dockerenv"); err == nil {
		return true
	}
	return false
}
