package giDevice

import (
	"fmt"
	"os"
	"os/signal"
	"testing"
)

var dev Device

func setupDevice(t *testing.T) {
	setupUsbmux(t)
	devices, err := um.Devices()
	if err != nil {
		t.Fatal(err)
	}

	if len(devices) == 0 {
		t.Fatal("No Device")
	}

	dev = devices[0]
}
func Test_device_ReadPairRecord(t *testing.T) {
	setupDevice(t)

	pairRecord, err := dev.ReadPairRecord()
	if err != nil {
		t.Fatal(err)
	}

	t.Log(pairRecord.HostID, pairRecord.SystemBUID, pairRecord.WiFiMACAddress)
}

func Test_device_NewConnect(t *testing.T) {
	setupDevice(t)

	if _, err := dev.NewConnect(LockdownPort); err != nil {
		t.Fatal(err)
	}
}

func Test_device_DeletePairRecord(t *testing.T) {
	setupDevice(t)

	if err := dev.DeletePairRecord(); err != nil {
		t.Fatal(err)
	}

}

func Test_device_SavePairRecord(t *testing.T) {
	setupLockdownSrv(t)

	pairRecord, err := lockdownSrv.Pair()
	if err != nil {
		t.Fatal(err)
	}

	err = dev.SavePairRecord(pairRecord)
	if err != nil {
		t.Fatal(err)
	}
}

func Test_device_XCTest(t *testing.T) {
	setupLockdownSrv(t)

	bundleID = "com.leixipaopao.WebDriverAgentRunner.xctrunner"
	out, cancel, err := dev.XCTest(bundleID)
	if err != nil {
		t.Fatal(err)
	}

	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt)

	go func() {
		for s := range out {
			fmt.Print(s)
		}
		done <- os.Interrupt
	}()

	for {
		select {
		case <-done:
			cancel()
			fmt.Println()
			t.Log("DONE")
			return
		}
	}
}
