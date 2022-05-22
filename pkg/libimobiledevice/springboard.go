package libimobiledevice

const (
	SpringBoardServiceName = "com.apple.springboardservices"
)

type SpringBoardBasicRequest struct {
	Label         string      `plist:"Label"`
	Command       CommandType `plist:"Command"`
	FormatVersion string      `plist:"FormatVersion"`
	BundleId      string      `plist:"BundleId"`
	IconState     string      `plist:"IconState"`
}

const (
	CommandTypeGetIconState   CommandType = "GetIconState"
	CommandTypeSetIconState   CommandType = "SetIconState"
	CommandTypeGetIconPNGData CommandType = "GetIconPNGData"
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

func (c *SpringBoardClient) NewBasicRequest(command CommandType) *SpringBoardBasicRequest {
	return &SpringBoardBasicRequest{
		Command: command,
		Label:   BundleID,
	}
}

func (c *SpringBoardClient) NewXmlPacket(req interface{}) (Packet, error) {
	return c.client.NewXmlPacket(req)
}

func (c *SpringBoardClient) SendPacket(pkt Packet) (err error) {
	return c.client.SendPacket(pkt)
}

func (c *SpringBoardClient) ReceivePacket() (respPkt Packet, err error) {
	return c.client.ReceivePacket()
}
