package configfile

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"net"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/Notifiarr/notifiarr/pkg/mnd"
)

/* This is a helper method to check if an IP is in a list/cidr. */

// AllowedIPs determines who can set x-forwarded-for.
type AllowedIPs struct {
	Input []string
	Nets  []*net.IPNet
}

var _ = fmt.Stringer(AllowedIPs{})

// String turns a list of allowedIPs into a printable masterpiece.
func (n AllowedIPs) String() string {
	if len(n.Nets) < 1 {
		return "(none)"
	}

	var output string

	for i := range n.Nets {
		if output != "" {
			output += ", "
		}

		output += n.Nets[i].String()
	}

	return output
}

// Contains returns true if an IP is allowed.
func (n AllowedIPs) Contains(ip string) bool {
	ip = strings.Trim(ip[:strings.LastIndex(ip, ":")], "[]")

	for i := range n.Nets {
		if n.Nets[i].Contains(net.ParseIP(ip)) {
			return true
		}
	}

	return false
}

// MakeIPs turns a list of CIDR strings (or plain IPs) into a list of net.IPNet.
// This "allowed" list is later used to check incoming IPs from web requests.
func MakeIPs(upstreams []string) AllowedIPs {
	allowed := AllowedIPs{
		Input: make([]string, len(upstreams)),
		Nets:  []*net.IPNet{},
	}

	for idx, ipAddr := range upstreams {
		allowed.Input[idx] = ipAddr

		if !strings.Contains(ipAddr, "/") {
			if strings.Contains(ipAddr, ":") {
				ipAddr += "/128"
			} else {
				ipAddr += "/32"
			}
		}

		if _, i, err := net.ParseCIDR(ipAddr); err == nil {
			allowed.Nets = append(allowed.Nets, i)
		}
	}

	return allowed
}

// CheckPort attempts to bind to a port to check if it's in use or not.
// We use this to check the port before starting the webserver.
func CheckPort(addr string) (string, error) {
	// Cleanup user input.
	addr = strings.TrimPrefix(strings.TrimPrefix(strings.TrimRight(addr, "/"), "http://"), "https://")
	if addr == "" {
		addr = mnd.DefaultBindAddr
	} else if !strings.Contains(addr, ":") {
		addr = "0.0.0.0:" + addr
	}

	a, err := net.ResolveTCPAddr("tcp", addr)
	if err != nil {
		return addr, fmt.Errorf("provided ip:port combo is invalid: %w", err)
	}

	l, err := net.ListenTCP("tcp", a)
	if err != nil {
		return addr, fmt.Errorf("unable to listen on provided ip:port: %w", err)
	}
	defer l.Close()

	return addr, nil
}

// BackupFile makes a config file backup file.
func BackupFile(configFile string) error {
	date := time.Now().Format("20060102T150405") // for file names.
	backupDir := filepath.Join(filepath.Dir(configFile), "backups")

	if err := os.MkdirAll(backupDir, mnd.Mode0755); err != nil {
		return fmt.Errorf("making config backup directory: %w", err)
	}

	deleteOldBackups(backupDir)

	// make config file backup.
	bckupFile := filepath.Join(backupDir, "backup.notifiarr."+date+".conf")
	if err := copyFile(configFile, bckupFile); err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			return fmt.Errorf("backing up config file: %w", err)
		}
	}

	return nil
}

func deleteOldBackups(backupDir string) {
	oldFiles, _ := filepath.Glob(filepath.Join(backupDir, "backup.notifiarr.*.conf"))
	fileList := []fs.FileInfo{}

	for _, file := range oldFiles {
		fileStat, err := os.Stat(file)
		if err != nil || fileStat.IsDir() {
			continue
		}

		if len(fileList) == 0 {
			fileList = []fs.FileInfo{fileStat} // first item
			continue
		} else if fileList[len(fileList)-1].ModTime().Before(fileStat.ModTime()) {
			fileList = append(fileList, fileStat) // last item
			continue
		}

		for idx, loopFile := range fileList { // somewhere in the middle.
			if loopFile.ModTime().After(fileStat.ModTime()) {
				fileList = append(fileList[:idx], append([]fs.FileInfo{fileStat}, fileList[idx:]...)...)
				break
			}
		}
	}

	// Keep newest 10 files.
	for i := range len(fileList) - 9 {
		os.Remove(filepath.Join(backupDir, fileList[i].Name()))
	}
}

// copyFile can be used to make a config file backup.
func copyFile(src, dst string) error {
	if _, err := os.Stat(dst); err == nil {
		return fmt.Errorf("%w: cannot overwrite file: %s", os.ErrExist, dst)
	}

	srcFile, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("opening existing file: %w", err)
	}
	defer srcFile.Close()

	srcStat, err := srcFile.Stat()
	if err != nil {
		return fmt.Errorf("stating existing file: %w", err)
	}

	dstFile, err := os.OpenFile(dst, os.O_CREATE|os.O_TRUNC|os.O_RDWR, srcStat.Mode())
	if err != nil {
		return fmt.Errorf("creating new file: %w", err)
	}
	defer dstFile.Close()

	if _, err = io.Copy(dstFile, srcFile); err != nil {
		return fmt.Errorf("copying data: %w", err)
	}

	return nil
}
