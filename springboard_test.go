package giDevice

import (
	"fmt"
	"testing"
)

var springBoardSrv SpringBoard

func setupSpringBoardSrv(t *testing.T) {
	setupLockdownSrv(t)

	var err error
	if lockdownSrv, err = dev.lockdownService(); err != nil {
		t.Fatal(err)
	}

	if springBoardSrv, err = lockdownSrv.SpringBoardService(); err != nil {
		t.Fatal(err)
	}
}

func Test_springBoard(t *testing.T) {
	setupSpringBoardSrv(t)
	fmt.Println(springBoardSrv.GetIconPNGData())
}
