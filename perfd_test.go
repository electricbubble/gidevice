package giDevice

import (
	"fmt"
	"testing"
	"time"
)

func TestPerfSystemMonitor(t *testing.T) {
	setupLockdownSrv(t)

	data, err := dev.PerfStart(
		WithPerfSystemCPU(true),
		WithPerfSystemMem(true),
		WithPerfSystemDisk(true),
		WithPerfSystemNetwork(true),
		WithPerfOutputInterval(1000),
	)
	if err != nil {
		t.Fatal(err)
	}

	timer := time.NewTimer(time.Duration(time.Second * 10))
	for {
		select {
		case <-timer.C:
			dev.PerfStop()
			return
		case d := <-data:
			fmt.Println(string(d))
		}
	}
}

func TestPerfProcessMonitor(t *testing.T) {
	setupLockdownSrv(t)

	data, err := dev.PerfStart(
		WithPerfProcessAttributes("pid", "cpuUsage", "memAnon"),
		WithPerfOutputInterval(1000),
		WithPerfPID(100),
	)
	if err != nil {
		t.Fatal(err)
	}

	timer := time.NewTimer(time.Duration(time.Second * 10))
	for {
		select {
		case <-timer.C:
			dev.PerfStop()
			return
		case d := <-data:
			fmt.Println(string(d))
		}
	}
}

func TestPerfGPU(t *testing.T) {
	setupLockdownSrv(t)

	data, err := dev.PerfStart(
		WithPerfSystemCPU(false),
		WithPerfSystemMem(false),
		WithPerfGPU(true),
	)
	if err != nil {
		t.Fatal(err)
	}

	timer := time.NewTimer(time.Duration(time.Second * 10))
	for {
		select {
		case <-timer.C:
			dev.PerfStop()
			return
		case d := <-data:
			fmt.Println(string(d))
		}
	}
}

func TestPerfFPS(t *testing.T) {
	setupLockdownSrv(t)

	data, err := dev.PerfStart(
		WithPerfSystemCPU(false),
		WithPerfSystemMem(false),
		WithPerfFPS(true),
	)
	if err != nil {
		t.Fatal(err)
	}

	timer := time.NewTimer(time.Duration(time.Second * 10))
	for {
		select {
		case <-timer.C:
			dev.PerfStop()
			return
		case d := <-data:
			fmt.Println(string(d))
		}
	}
}

func TestPerfNetwork(t *testing.T) {
	setupLockdownSrv(t)

	data, err := dev.PerfStart(
		WithPerfSystemCPU(false),
		WithPerfSystemMem(false),
		WithPerfNetwork(true),
	)
	if err != nil {
		t.Fatal(err)
	}

	timer := time.NewTimer(time.Duration(time.Second * 10))
	for {
		select {
		case <-timer.C:
			dev.PerfStop()
			return
		case d := <-data:
			fmt.Println(string(d))
		}
	}
}
