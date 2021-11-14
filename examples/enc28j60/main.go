package main

import (
	"machine"
	"net"
	"time"

	"tinygo.org/x/drivers/enc28j60"
)

// Example for Raspberry Pi Pico
const (
	csPin  = machine.GPIO5
	sckPin = machine.GPIO2
	sdoPin = machine.GPIO3
	sdiPin = machine.GPIO4
)

var (
	spi = machine.SPI0
	err error
	// Max eth
	buffer [enc28j60.MAX_FRAMELEN]byte
	MAC    = net.HardwareAddr{0, 0, 0, 0, 0, 0}
)

func main() {
	csPin.Configure(machine.PinConfig{Mode: machine.PinOutput})
	csPin.High()
	println("SP") // Start Program
	err = spi.Configure(machine.SPIConfig{
		Frequency: 0.5e6, // Datasheet recommends at least 8MHz
		SCK:       sckPin,
		SDO:       sdoPin,
		SDI:       sdiPin,
	})
	if err != nil {
		panic(err.Error())
	}

	e := enc28j60.New(csPin, spi)
	err = e.Init(MAC)
	if err != nil {
		panic(err.Error())
	}
	println("chip revision:", e.GetRev())
	// Packet length
	var plen uint16
	for {
		// see if packet has been recieved
		plen = e.PacketRecieve(buffer[:])
		if plen != 0 {
			println(string(buffer[:plen]))
			// e.PacketSend(buffer[:plen])
		}
		time.Sleep(3 * time.Second)
	}
}
