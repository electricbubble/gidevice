package giDevice

import (
	"testing"
	"time"
)

func TestPerfCPUMem(t *testing.T) {
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

func TestPerfGPU(t *testing.T) {
	setupLockdownSrv(t)

	data, err := dev.PerfStart(
		WithPerfCPU(false),
		WithPerfMem(false),
		WithPerfGPU(true),
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

func TestPerfFPS(t *testing.T) {
	setupLockdownSrv(t)

	data, err := dev.PerfStart(
		WithPerfCPU(false),
		WithPerfMem(false),
		WithPerfFPS(true),
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

func TestPerfNetwork(t *testing.T) {
	setupLockdownSrv(t)

	data, err := dev.PerfStart(
		WithPerfCPU(false),
		WithPerfMem(false),
		WithPerfNetwork(true),
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
