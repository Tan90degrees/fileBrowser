package server

import (
	"errors"
	"log"
	"path/filepath"
	"strings"
	"syscall"
)

func checkPath(root string, servPath string) error {
	rootAbs, err := filepath.Abs(root)
	if err != nil {
		log.Println(err)
		return errors.New("illegal address")
	}
	servPathAbs, err := filepath.Abs(servPath)
	if err != nil {
		log.Println(err)
		return errors.New("illegal address")
	}
	if !strings.HasPrefix(servPathAbs, rootAbs) {
		return errors.New("illegal address")
	}
	if (len(servPathAbs) > len(rootAbs)) && (servPathAbs[len(rootAbs)] != '/') {
		return errors.New("illegal address")
	}
	return nil
}

func diskCapacityEnough(path string, size uint64) bool {
	fs := syscall.Statfs_t{}
	err := syscall.Statfs(path, &fs)
	if err != nil {
		return false
	}
	freeBytes := fs.Bavail * uint64(fs.Bsize)
	return freeBytes >= size
}
