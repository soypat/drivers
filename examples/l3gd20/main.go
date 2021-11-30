package main

import (
	"machine"
	"time"

	"tinygo.org/x/drivers/l3gd20"
)

func main() {
	const (
		// Default address on most breakout boards.
		pcaAddr = 0x40
	)
	led := machine.LED
	led.Configure(machine.PinConfig{Mode: machine.PinOutput})
	bus := machine.I2C1
	err := bus.Configure(machine.I2CConfig{})
	if err != nil {
		panic(err.Error())
	}
	gyro := l3gd20.NewI2C(bus, 105)
	err = gyro.Configure(l3gd20.Config{Range: l3gd20.Range_250})
	if err != nil {
		addrs := ScanI2CDev(bus)
		println("found addresses:")
		for _, addr := range addrs {
			println(addr)
		}
		led.High()
		panic(err.Error())
	}
	var x, y, z int32
	for {
		err = gyro.Update()
		if err != nil {
			println(err.Error())
		}
		x, y, z = gyro.AngularVelocity()
		println(x, y, z)
		led.High()
		time.Sleep(250 * time.Millisecond)
		led.Low()
		time.Sleep(250 * time.Millisecond)
	}
}

// ScanI2CDev finds I2C devices on the bus and rreturns them inside
// a slice. If slice is nil then no devices were found.
func ScanI2CDev(bus *machine.I2C) (addrs []uint8) {
	var addr, count uint8
	var err error
	w := []byte{1}
	// Count devices in first scan
	for addr = 1; addr < 127; addr++ {
		err = bus.Tx(uint16(addr), w, nil)
		if err == nil {
			count++
		}
	}
	if count == 0 {
		return nil
	}
	// Allocate slice and populate slice with addresses
	addrs = make([]uint8, count)
	count = 0
	for addr = 1; addr < 127; addr++ {
		err = bus.Tx(uint16(addr), w, nil)
		if err == nil && count < uint8(len(addrs)) {
			addrs[count] = addr
			count++
		}
	}
	return addrs
}
