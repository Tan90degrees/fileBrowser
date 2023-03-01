package server

import (
	"errors"
	"io/fs"
	"log"
	"path/filepath"
	"strings"
	"syscall"
	"unsafe"

	"golang.org/x/sys/windows"
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
	if (len(servPathAbs) > len(rootAbs)) && (servPathAbs[len(rootAbs)] != '\\') {
		return errors.New("illegal address")
	}
	return nil
}

func diskCapacityEnough(lpDirectoryName string, size uint64) bool {
	k32dll := syscall.MustLoadDLL("kernel32.dll")
	defer k32dll.Release()
	proc := k32dll.MustFindProc("GetDiskFreeSpaceExW")
	lpFreeBytesAvailableToCaller := uint64(0)
	_, _, err := proc.Call(uintptr(unsafe.Pointer(windows.StringToUTF16Ptr(lpDirectoryName))),
		uintptr(unsafe.Pointer(&lpFreeBytesAvailableToCaller)))
	if errors.Is(err, fs.ErrNotExist) {
		return false
	}
	return lpFreeBytesAvailableToCaller >= size
}
