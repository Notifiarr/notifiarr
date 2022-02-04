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
		return "error with syscall"
	}

	userName, _ := user.LookupId(strconv.FormatUint(uint64(stat.Uid), mnd.Base10))
	groupName, _ := user.LookupGroupId(strconv.FormatUint(uint64(stat.Gid), mnd.Base10))

	return userName.Name + ", " + userName.Username + ":" + groupName.Name
}
