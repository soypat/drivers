package main

import (
	"machine"
	"time"

	"tinygo.org/x/drivers/lora"
	"tinygo.org/x/drivers/sx127x"
)

func main() {
	time.Sleep(time.Second)
	println("start program")
	defer println("end program")
	spi := machine.SPI0

	err := spi.Configure(machine.SPIConfig{
		Frequency: 100000,
		SCK:       machine.GP2,
		SDO:       machine.GP3,
		SDI:       machine.GP4,
	})
	if err != nil {
		panic(err)
	}
	const (
		cs   = machine.GP8
		rst  = machine.GP6
		dio0 = machine.GP7
	)
	rst.Configure(machine.PinConfig{Mode: machine.PinOutput})
	cs.Configure(machine.PinConfig{Mode: machine.PinOutput})
	cs.High()
	toodledoo := sx127x.New(spi, rst, cs)
	toodledoo.Reset()
	if !toodledoo.DetectDevice() {
		panic("device not detected")
	}
	toodledoo.LoraConfig(lora.Config{
		Crc:            lora.CRCOn,
		Iq:             lora.IQStandard,
		SyncWord:       lora.SyncPublic,
		HeaderType:     lora.HeaderExplicit,
		Freq:           433500000,
		Cr:             lora.CodingRate4_7,
		Sf:             lora.SpreadingFactor9,
		Bw:             lora.Bandwidth_125_0,
		Ldr:            lora.LowDataRateOptimizeOn,
		Preamble:       8,
		LoraTxPowerDBm: 20,
	})
	toodledoo.SetRxTimeout(255)
	// RX RX RX RX RX RX RX RX RX
	for {
		startRx := time.Now()
		packet, err := toodledoo.Rx(7000)
		if err != nil {
			println("error receiving message", err.Error())
		} else {
			println("message rx in", time.Since(startRx).String())
			println(string(packet))
		}
	}
}
