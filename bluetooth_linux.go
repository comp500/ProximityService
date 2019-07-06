// +build linux

package main

import (
	"fmt"
	"log"
	"strings"

	"github.com/paypal/gatt"
)

func startBluetooth(dataChannel chan []byte, done chan bool) {
	d, err := gatt.NewDevice(gatt.LnxMaxConnections(1), gatt.LnxDeviceID(-1, true))
	if err != nil {
		log.Fatalf("Failed to open device, err: %s\n", err)
		return
	}

	handler := gattHandler{dataChannel, done}

	// Register handlers.
	d.Handle(
		gatt.PeripheralDiscovered(onPeriphDiscovered),
		gatt.PeripheralConnected(handler.onPeriphConnected),
		gatt.PeripheralDisconnected(onPeriphDisconnected),
	)

	d.Init(onStateChanged)

	<-done
	d.StopScanning()
}

type gattHandler struct {
	dataChannel chan []byte
	done        chan bool
}

func onStateChanged(d gatt.Device, s gatt.State) {
	fmt.Println("State:", s)
	switch s {
	case gatt.StatePoweredOn:
		fmt.Println("Scanning...")
		d.Scan([]gatt.UUID{}, false)
		return
	default:
		d.StopScanning()
	}
}

func onPeriphDiscovered(p gatt.Peripheral, a *gatt.Advertisement, rssi int) {
	// TODO: make this configurable
	if strings.ToUpper(p.ID()) != "E2:85:89:64:5C:84" {
		return
	}

	// Stop scanning once we've got the peripheral we're looking for.
	p.Device().StopScanning()

	fmt.Printf("\nPeripheral ID:%s, NAME:(%s)\n", p.ID(), p.Name())
	fmt.Println("  Local Name        =", a.LocalName)
	fmt.Println("  TX Power Level    =", a.TxPowerLevel)
	fmt.Println("  Manufacturer Data =", a.ManufacturerData)
	fmt.Println("  Service Data      =", a.ServiceData)
	fmt.Println("")

	p.Device().Connect(p)
}

func (g gattHandler) onPeriphConnected(p gatt.Peripheral, err error) {
	fmt.Println("Connected")

	if err := p.SetMTU(500); err != nil {
		fmt.Printf("Failed to set MTU, err: %s\n", err)
	}

	// Nordic UART service
	srvUUID := gatt.MustParseUUID("6E400001-B5A3-F393-E0A9-E50E24DCCA9E")
	txdUUID := gatt.MustParseUUID("6E400002-B5A3-F393-E0A9-E50E24DCCA9E")
	rxdUUID := gatt.MustParseUUID("6E400003-B5A3-F393-E0A9-E50E24DCCA9E")
	txdFound := false
	rxdFound := false

	// Discover services
	ss, err := p.DiscoverServices(nil)
	if err != nil {
		fmt.Printf("Failed to discover services, err: %s\n", err)
		p.Device().CancelConnection(p)
		return
	}

	for _, s := range ss {
		if !s.UUID().Equal(srvUUID) {
			continue
		}

		fmt.Println("UART service found")

		// Discover characteristics
		cs, err := p.DiscoverCharacteristics(nil, s)
		if err != nil {
			fmt.Printf("Failed to discover characteristics, err: %s\n", err)
			continue
		}

		for _, c := range cs {
			if c.UUID().Equal(txdUUID) {
				txdFound = true
				fmt.Println("TXD characteristic found")

			} else if c.UUID().Equal(rxdUUID) {
				rxdFound = true
				fmt.Println("RXD characteristic found, subscribing...")

				// For some reason, DiscoverDescriptors must be called before subscribing
				_, _ = p.DiscoverDescriptors(nil, c)

				f := func(c *gatt.Characteristic, b []byte, err error) {
					g.dataChannel <- b
				}
				if err := p.SetNotifyValue(c, f); err != nil {
					fmt.Printf("Failed to subscribe characteristic, err: %s\n", err)
					continue
				}
			} else {
				continue
			}
		}
	}

	if !txdFound || !rxdFound {
		fmt.Println("Service not found, disconnecting...")
		p.Device().CancelConnection(p)
	}

	<-g.done
	p.Device().CancelConnection(p)
}

func onPeriphDisconnected(p gatt.Peripheral, err error) {
	fmt.Println("Disconnected, scanning...")
	p.Device().Scan([]gatt.UUID{}, false)
}
