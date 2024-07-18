//go:build !windows

package logs

import (
	"os"
	"os/user"
	"strconv"
	"syscall"

	"github.com/Notifiarr/notifiarr/pkg/mnd"
)

func getFileOwner(fileInfo os.FileInfo) string {
	stat, ok := fileInfo.Sys().(*syscall.Stat_t)
	if !ok {
		return "error with syscall; might a Notifiarr bug, please report!"
	}

	uid := strconv.FormatUint(uint64(stat.Uid), mnd.Base10)
	gid := strconv.FormatUint(uint64(stat.Gid), mnd.Base10)
	name := ""
	usr := ""
	grp := ""

	if userName, err := user.LookupId(uid); err == nil {
		usr = userName.Username
		name = userName.Name
	}

	if groupName, err := user.LookupGroupId(gid); err == nil {
		grp = groupName.Name
	}

	if usr != "" {
		return name + ", " + usr + ":" + grp + " (" + uid + ":" + gid + ")"
	}

	return uid + ":" + gid
}

func hasConsoleWindow() bool {
	return true
}
