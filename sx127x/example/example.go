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
		Freq:           433500000,
		Cr:             lora.CodingRate4_7,
		Sf:             lora.SpreadingFactor9,
		Bw:             lora.Bandwidth_125_0,
		Ldr:            lora.LowDataRateOptimizeOn,
		Preamble:       8,
		LoraTxPowerDBm: 20,
	})
	for {
		err = toodledoo.Tx([]byte("Hello World!"), 10000)
		if err != nil {
			println("error sending message", err.Error())
		} else {
			println("message sent")
		}
		time.Sleep(5 * time.Second)
	}
}
