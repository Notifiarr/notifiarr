package client

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"maps"
	"net"
	"os"
	"strings"

	"github.com/Notifiarr/notifiarr/pkg/snapshot"
	"github.com/jackpal/gateway"
)

func environ() map[string]string {
	out := make(map[string]string)

	for _, v := range os.Environ() {
		if s := strings.SplitN(v, "=", 2); len(s) == 2 && s[0] != "" { //nolint:mnd
			out[s[0]] = s[1]
		}
	}

	return out
}

// getLinesFromFile makes it easy to tail or head a file. Sorta.
func getLinesFromFile(filepath, sort string, count, skip int) ([]byte, error) {
	fileHandle, err := os.Open(filepath)
	if err != nil {
		return nil, fmt.Errorf("opening file: %w", err)
	}
	defer fileHandle.Close()

	stat, err := fileHandle.Stat()
	if err != nil {
		return nil, fmt.Errorf("stating open file: %w", err)
	}

	switch sort {
	default:
		fallthrough
	case "tail", "tails":
		return readFileTail(fileHandle, stat.Size(), count, skip)
	case "head", "heads":
		return readFileHead(fileHandle, stat.Size(), count, skip)
	}
}

func readFileTail(fileHandle *os.File, fileSize int64, count, skip int) ([]byte, error) { //nolint:cyclop
	var (
		output   bytes.Buffer
		location int64
		filesize = fileSize
		char     = make([]byte, 1)
		found    int
	)

	// This is a magic number.
	// We assume 150 characters per line to optimize the buffer.
	output.Grow(count * 150) //nolint:mnd

	for {
		location-- // read 1 byte
		if _, err := fileHandle.Seek(location, io.SeekEnd); err != nil {
			return nil, fmt.Errorf("seeking open file: %w", err)
		}

		if _, err := fileHandle.Read(char); err != nil {
			return nil, fmt.Errorf("reading open file: %w", err)
		}

		if location != -1 && (char[0] == 10) { //nolint:mnd
			found++ // we found a line
		}

		if skip == 0 || found >= skip {
			output.WriteByte(char[0])
		}

		if found >= count+skip || // we found enough lines.
			location == -filesize { // beginning of file.
			out := revBytes(output)
			if len(out) > 0 && out[0] == '\n' {
				return out[1:], nil // strip off the /n
			}

			return out, nil
		}
	}
}

func readFileHead(fileHandle *os.File, fileSize int64, count, skip int) ([]byte, error) {
	var (
		output   bytes.Buffer
		location int64
		char     = make([]byte, 1)
		found    int
	)

	// This is a magic number.
	// We assume 150 characters per line to optimize the buffer.
	output.Grow(count * 150) //nolint:mnd

	for ; ; location++ {
		if _, err := fileHandle.Seek(location, io.SeekStart); err != nil {
			return nil, fmt.Errorf("seeking open file: %w", err)
		}

		if _, err := fileHandle.Read(char); err != nil {
			return nil, fmt.Errorf("reading open file: %w", err)
		}

		if char[0] == 10 { //nolint:mnd
			found++ // we have a line

			if found <= skip {
				// skip writing new lines until we get to our first line.
				continue
			}
		}

		if found >= skip {
			output.WriteByte(char[0])
		}

		if found >= count+skip || // we found enough lines.
			location >= fileSize-1 { // end of file.
			return output.Bytes(), nil
		}
	}
}

// revBytes returns a bytes buffer reversed.
func revBytes(output bytes.Buffer) []byte {
	data := output.Bytes()
	for i, j := 0, len(data)-1; i < j; i, j = i+1, j-1 {
		data[i], data[j] = data[j], data[i]
	}

	return data
}

func getDisks(ctx context.Context, zfsPools []string) map[string]*snapshot.Partition {
	disks, _ := snapshot.GetDisksUsage(ctx, true)
	zfspools, _ := snapshot.GetZFSPoolData(ctx, zfsPools)
	output := make(map[string]*snapshot.Partition)

	maps.Copy(output, disks)
	maps.Copy(output, zfspools)

	return output
}

func getGateway() string {
	gateway, err := gateway.DiscoverGateway()
	if err != nil {
		return ""
	}

	return gateway.String()
}

// Returns interface name and netmask.
func getIfNameAndNetmask(ipAddr string) (string, string) {
	ifaces, err := net.Interfaces()
	if err != nil {
		return "", ""
	}

	for _, iface := range ifaces {
		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}

		for _, addr := range addrs {
			switch address := addr.(type) {
			case *net.IPNet:
				if address.IP.String() == ipAddr {
					return iface.Name, addr.String()
				}
			case *net.IPAddr:
				if address.IP.String() == ipAddr {
					return iface.Name, addr.String()
				}
			}
		}
	}

	return "", ""
}
