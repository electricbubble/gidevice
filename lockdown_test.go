package giDevice

import (
	"testing"
)

var lockdownSrv Lockdown

func setupLockdownSrv(t *testing.T) {
	setupDevice(t)

	var err error
	if lockdownSrv, err = dev.lockdownService(); err != nil {
		t.Fatal(err)
	}
}

func Test_lockdown_QueryType(t *testing.T) {
	setupLockdownSrv(t)

	lockdownType, err := lockdownSrv.QueryType()
	if err != nil {
		t.Fatal(err)
	}

	t.Log(lockdownType.Type)
}

func Test_lockdown_GetValue(t *testing.T) {
	setupLockdownSrv(t)

	v, err := lockdownSrv.GetValue("", "")
	// v, err := lockdownSrv.GetValue("", "ProductVersion")
	// v, err := lockdownSrv.GetValue("", "DeviceName")
	// v, err := lockdownSrv.GetValue("com.apple.mobile.iTunes", "")
	// v, err := lockdownSrv.GetValue("com.apple.mobile.battery", "")
	// v, err := lockdownSrv.GetValue("com.apple.disk_usage", "")
	if err != nil {
		t.Fatal(err)
	}

	t.Log(v)
}
