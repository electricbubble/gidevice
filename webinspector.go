package giDevice

import "github.com/electricbubble/gidevice/pkg/libimobiledevice"

func newWebInspector(client *libimobiledevice.WebInspectClient) *webInspector {
	return &webInspector{
		client: client,
	}
}

type webInspector struct {
	client *libimobiledevice.WebInspectClient
}
