package giDevice

import (
	"testing"
	"time"
)

func TestPerf(t *testing.T) {
	setupLockdownSrv(t)

	data, err := dev.PerfStart(
		WithPerfCPU(true),
		WithPerfMem(true),
		WithPerfBundleID("com.apple.Spotlight"),
	)
	if err != nil {
		t.Fatal(err)
	}

	timer := time.NewTimer(time.Duration(time.Second * 10))
	for {
		select {
		case <-timer.C:
			return
		case d := <-data:
			t.Log(string(d))
		}
	}
}
