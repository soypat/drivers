package main

import (
	"machine"
	"time"

	"tinygo.org/x/drivers/pca9685"
)

func main() {
	const (
		// Default address on most breakout boards.
		pcaAddr = 0x40
		// 200Hz PWM
		period = 1e9 / 200
	)
	err := machine.I2C0.Configure(machine.I2CConfig{})
	if err != nil {
		panic(err.Error())
	}
	d := pca9685.New(machine.I2C0, pcaAddr)
	err = d.IsConnected()
	if err != nil {
		panic(err.Error())
	}
	err = d.Configure(pca9685.PWMConfig{Period: period})
	if err != nil {
		panic(err.Error())
	}
	var value uint32
	step := d.Top() / 5
	for count := 0; count < 10; count++ {
		for value = 0; value <= d.Top(); value += step {
			d.SetAll(value)
			dc := 100 * value / d.Top()
			println("set dc @", dc, "%")
			time.Sleep(800 * time.Millisecond)
		}
	}

	// Below is the usage of DevBuffered which is well suited for fast response
	// PWM signals.
	db := pca9685.NewBuffered(machine.I2C0, pcaAddr)
	err = db.IsConnected()
	if err != nil {
		panic(err.Error())
	}
	err = db.Configure(pca9685.PWMConfig{Period: period})
	if err != nil {
		panic(err.Error())
	}
	for ch := 0; ch < 16; ch++ {
		// This does not perform IO on I2C bus.
		db.PrepSet(0, db.Top()/uint32(ch+1))
	}
	// All 16 PWM registers are written in one I2C transaction
	// to minimize bus overhead.
	err = db.Update()
	if err != nil {
		panic(err.Error())
	}
}

// ScanI2CDev finds I2C devices on the bus and rreturns them inside
// a slice. If slice is nil then no devices were found.
func ScanI2CDev(bus machine.I2C) (addrs []uint8) {
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
