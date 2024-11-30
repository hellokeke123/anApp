//go:build windows

package winsys

import (
	"golang.org/x/sys/windows"
	"os"
	"unsafe"
)

func CreateDisplayData(name, description string) FWPM_DISPLAY_DATA0 {
	namePtr, err := windows.UTF16PtrFromString(name)
	Must(err)

	descriptionPtr, err := windows.UTF16PtrFromString(description)
	Must(err)

	return FWPM_DISPLAY_DATA0{
		Name:        namePtr,
		Description: descriptionPtr,
	}
}

func GetCurrentProcessAppID() (*FWP_BYTE_BLOB, error) {
	currentFile, err := os.Executable()
	if err != nil {
		return nil, err
	}

	curFilePtr, err := windows.UTF16PtrFromString(currentFile)
	if err != nil {
		return nil, err
	}

	windows.GetCurrentProcessId()

	var appID *FWP_BYTE_BLOB
	err = FwpmGetAppIdFromFileName0(curFilePtr, unsafe.Pointer(&appID))
	if err != nil {
		return nil, err
	}
	return appID, nil
}

func Must(errs ...error) {
	for _, err := range errs {
		if err != nil {
			panic(err)
		}
	}
}
