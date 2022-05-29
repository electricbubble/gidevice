package libimobiledevice

const (
	WebInspectorServiceName = "com.apple.webinspector"
)

func NewWebInspectorClient(innerConn InnerConn) *WebInspectClient {
	return &WebInspectClient{
		newServicePacketClient(innerConn),
	}
}

type WebInspectClient struct {
	client *servicePacketClient
}

func (c *WebInspectClient) InnerConn() InnerConn {
	return c.client.innerConn
}

func (c *WebInspectClient) NewXmlPacket(req interface{}) (Packet, error) {
	return c.client.NewXmlPacket(req)
}

func (c *WebInspectClient) SendPacket(pkt Packet) (err error) {
	return c.client.SendPacket(pkt)
}

func (c *WebInspectClient) ReceivePacket() (respPkt Packet, err error) {
	return c.client.ReceivePacket()
}
