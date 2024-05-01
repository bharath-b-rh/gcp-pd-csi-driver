//go:build linux

/*
Copyright 2022 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at
    http://www.apache.org/licenses/LICENSE-2.0
Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package deviceutils

import (
	"fmt"
	"os"
	"path/filepath"

	"k8s.io/mount-utils"
)

// DisableDevice asks the kernel to disable a device via /sys.
// NB: this can be dangerous to use. Once a device is disabled, it's unusable, and can't be enabled
// unless the serial number is known. But the serial number cannot be read from the device as it's
// disabled. If a device is disabled in NodeUnstage, and then NodeStage is called without a
// NodeUnpublish & publish sequence, the disabled state of the device will cause NodeStage to fail.
// So this can only be used if we track the serial numbers of disabled devices in a persisent way
// that survives driver restarts.
func (_ *deviceUtils) DisableDevice(devicePath string) error {
	deviceName := filepath.Base(devicePath)
	return os.WriteFile(fmt.Sprintf("/sys/block/%s/device/state", deviceName), []byte("offline\n"), 0644)
}

func (_ *deviceUtils) IsDeviceFilesystemInUse(mounter *mount.SafeFormatAndMount, devicePath, devFsPath string) (bool, error) {
	fstype, err := mounter.GetDiskFormat(devicePath)
	if err != nil {
		return false, fmt.Errorf("failed to get disk format for %s (aka %s): %v", devicePath, devFsPath, err)
	}

	devFsName := filepath.Base(devFsPath)
	sysFsTypePath := fmt.Sprintf("/sys/fs/%s/%s", fstype, devFsName)
	stat, err := os.Stat(sysFsTypePath)
	if err != nil {
		if os.IsNotExist(err) {
			// Path doesn't exist, indicating the device is NOT in use
			return false, nil
		}
		return false, err
	}

	return stat.IsDir(), nil
}
