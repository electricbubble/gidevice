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
	req := map[string]interface{}{
		"command": "getIconState",
	}
	if pkt, err = s.client.NewBinaryPacket(req); err != nil {
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

func (s springboard) GetIconPNGData() (err error) {
	//var pkt libimobiledevice.Packet
	//r := s.client.NewBasicRequest(libimobiledevice.CommandTypeGetIconPNGData)
	//r.BundleId = "com.tencent.xin"
	//if pkt, err = s.client.NewXmlPacket(r); err != nil {
	//	return
	//}
	//if err = s.client.SendPacket(pkt); err != nil {
	//	return err
	//}
	//var respPkt libimobiledevice.Packet
	//if respPkt, err = s.client.ReceivePacket(); err != nil {
	//	return err
	//}
	//fmt.Println(respPkt)
	return
}
