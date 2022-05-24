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

func (s springboard) GetIconPNGData() (pngData []byte, err error) {
	var pkt libimobiledevice.Packet
	req := map[string]interface{}{
		"command":  "getIconPNGData",
		"bundleId": "com.tencent.xin",
	}
	if pkt, err = s.client.NewBinaryPacket(req); err != nil {
		return
	}
	if err = s.client.SendPacket(pkt); err != nil {
		return nil, err
	}
	var respPkt libimobiledevice.Packet
	if respPkt, err = s.client.ReceivePacket(); err != nil {
		return nil, err
	}
	var reply libimobiledevice.IconPNGDataResponse
	if err = respPkt.Unmarshal(&reply); err != nil {
		return nil, fmt.Errorf("receive packet: %w", err)
	}
	pngData = reply.PNGData
	return
}