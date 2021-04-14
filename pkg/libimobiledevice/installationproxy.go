package libimobiledevice

const InstallationProxyServiceName = "com.apple.mobile.installation_proxy"

const (
	CommandTypeBrowse CommandType = "Browse"
	CommandTypeLookup CommandType = "Lookup"
)

type ApplicationType string

const (
	ApplicationTypeSystem   ApplicationType = "System"
	ApplicationTypeUser     ApplicationType = "User"
	ApplicationTypeInternal ApplicationType = "internal"
	ApplicationTypeAny      ApplicationType = "Any"
)

func NewInstallationProxyClient(innerConn InnerConn) *InstallationProxyClient {
	return &InstallationProxyClient{
		client: newServicePacketClient(innerConn),
	}
}

type InstallationProxyClient struct {
	client *servicePacketClient
}

func (c *InstallationProxyClient) NewBasicRequest(cmdType CommandType, opt *InstallationProxyOption) *InstallationProxyBasicRequest {
	req := &InstallationProxyBasicRequest{Command: cmdType}
	if opt != nil {
		req.ClientOptions = opt
	}
	return req
}

func (c *InstallationProxyClient) NewXmlPacket(req interface{}) (Packet, error) {
	return c.client.NewXmlPacket(req)
}

func (c *InstallationProxyClient) SendPacket(pkt Packet) (err error) {
	return c.client.SendPacket(pkt)
}

func (c *InstallationProxyClient) ReceivePacket() (respPkt Packet, err error) {
	return c.client.ReceivePacket()
}

type InstallationProxyOption struct {
	ApplicationType  ApplicationType `plist:"ApplicationType,omitempty"`
	ReturnAttributes []string        `plist:"ReturnAttributes,omitempty"`
	MetaData         bool            `plist:"com.apple.mobile_installation.metadata,omitempty"`
	BundleIDs        []string        `plist:"BundleIDs,omitempty"` /* for Lookup */
}

type (
	InstallationProxyBasicRequest struct {
		Command       CommandType              `plist:"Command"`
		ClientOptions *InstallationProxyOption `plist:"ClientOptions,omitempty"`
	}
)

type (
	InstallationProxyBasicResponse struct {
		Status string `plist:"Status"`
	}

	InstallationProxyLookupResponse struct {
		InstallationProxyBasicResponse
		LookupResult interface{} `plist:"LookupResult"`
	}

	InstallationProxyBrowseResponse struct {
		InstallationProxyBasicResponse
		CurrentAmount int           `json:"CurrentAmount"`
		CurrentIndex  int           `json:"CurrentIndex"`
		CurrentList   []interface{} `json:"CurrentList"`
	}
)
