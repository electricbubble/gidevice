# Golang-iDevice

## Installation

```shell script
go get github.com/electricbubble/gidevice
```

#### Devices

```go
package main

import (
	giDevice "github.com/electricbubble/gidevice"
	"log"
)

func main() {
	usbmux, err := giDevice.NewUsbmux()
	if err != nil {
		log.Fatalln(err)
	}

	devices, err := usbmux.Devices()
	if err != nil {
		log.Fatal(err)
	}

	for _, dev := range devices {
		log.Println(dev.Properties().SerialNumber, dev.Properties().ProductID, dev.Properties().DeviceID)
	}
}

```


#### XCTest

```go
package main

import (
	"fmt"
	giDevice "github.com/electricbubble/gidevice"
	"log"
	"sync"
	"time"
)

func main() {
	usbmux, err := giDevice.NewUsbmux()
	if err != nil {
		log.Fatal(err)
	}

	devices, err := usbmux.Devices()
	if err != nil {
		log.Fatal(err)
	}

	if len(devices) == 0 {
		log.Fatal("No Device")
	}

	dev := devices[0]

	out, cancel, err := dev.XCTest("com.leixipaopao.WebDriverAgentRunner.xctrunner")
	if err != nil {
		log.Fatal(err)
	}

	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		for s := range out {
			fmt.Print(s)
		}
	}()

	go func() {
		time.Sleep(10 * time.Second)
		cancel()
		log.Println("DONE")
		wg.Done()
	}()

	wg.Wait()
}

```

## Thanks

| |About|
|---|---|
|[libimobiledevice/libimobiledevice](https://github.com/libimobiledevice/libimobiledevice)|A cross-platform protocol library to communicate with iOS devices|
|[anonymous5l/iConsole](https://github.com/anonymous5l/iConsole)|iOS usbmuxd communication impl iTunes protocol|
|[alibaba/taobao-iphone-device](https://github.com/alibaba/taobao-iphone-device)|tidevice can be used to communicate with iPhone device|

Thank you [JetBrains](https://www.jetbrains.com/?from=gwda) for providing free open source licenses
