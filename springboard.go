package giDevice

import (
	"fmt"
	"github.com/electricbubble/gidevice/pkg/libimobiledevice"
)

func newSpringBoard(client *libimobiledevice.SpringBoardClient) *springboard {
	return &springboard{
		client: client,
	}
}

type springboard struct {
	client *libimobiledevice.SpringBoardClient
}

func (s springboard) GetIconState() (err error) {
	var pkt libimobiledevice.Packet
	if pkt, err = s.client.NewXmlPacket(
		s.client.NewBasicRequest(libimobiledevice.CommandTypeGetIconState),
	); err != nil {
		return
	}
	if err = s.client.SendPacket(pkt); err != nil {
		return err
	}
	var respPkt libimobiledevice.Packet
	if respPkt, err = s.client.ReceivePacket(); err != nil {
		return err
	}
	fmt.Println(respPkt)
	return
}

func (s springboard) SetIconState() {

}

func (s springboard) GetIconPNGData() {

}
