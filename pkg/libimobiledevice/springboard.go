package libimobiledevice

import (
	"bytes"
	"encoding/binary"
	"fmt"
)

type IconPNGDataResponse struct {
	PNGData []byte `plist:"pngData"`
}

const (
	SpringBoardServiceName = "com.apple.springboardservices"
)

func NewSpringBoardClient(innerConn InnerConn) *SpringBoardClient {
	return &SpringBoardClient{
		newServicePacketClient(innerConn),
	}
}

type SpringBoardClient struct {
	client *servicePacketClient
}

func (c *SpringBoardClient) InnerConn() InnerConn {
	return c.client.innerConn
}

func (c *SpringBoardClient) NewXmlPacket(req interface{}) (Packet, error) {
	return c.client.NewXmlPacket(req)
}

func (c *SpringBoardClient) SendPacket(pkt Packet) (err error) {
	return c.client.SendPacket(pkt)
}

func (c *SpringBoardClient) ReceivePacket() (respPkt Packet, err error) {
	var bufLen []byte
	if bufLen, err = c.client.innerConn.Read(4); err != nil {
		return nil, fmt.Errorf("receive packet: %w", err)
	}
	lenPkg := binary.BigEndian.Uint32(bufLen)

	buffer := bytes.NewBuffer([]byte{})
	buffer.Write(bufLen)

	var buf []byte
	if buf, err = c.client.innerConn.Read(int(lenPkg)); err != nil {
		return nil, fmt.Errorf("receive packet: %w", err)
	}
	buffer.Write(buf)

	if respPkt, err = new(servicePacket).Unpack(buffer); err != nil {
		return nil, fmt.Errorf("receive packet: %w", err)
	}

	debugLog(fmt.Sprintf("<-- %s\n", respPkt))

	return
}

func (c *SpringBoardClient) NewBinaryPacket(req interface{}) (Packet, error) {
	return c.client.NewBinaryPacket(req)
}
