package main

import (
	"machine"
	"time"

	"tinygo.org/x/drivers/lsm303dlh"
)

func main() {
	const (
		// Default address on most breakout boards.
		pcaAddr = 0x40
	)
	bus := machine.I2C0
	err := bus.Configure(machine.I2CConfig{})
	if err != nil {
		panic(err.Error())
	}
	d := lsm303dlh.New(machine.I2C0, lsm303dlh.I2CAddr)
	err = d.Configure(lsm303dlh.Config{
		MagGain: lsm303dlh.MAGGAIN_1_3,
		MagRate: lsm303dlh.MAGRATE_7_5,
	})
	if err != nil {
		println(err.Error())
		addrs := ScanI2CDev(bus)
		for _, addr := range addrs {
			println("found addr:", addr)
		}
	}
	if !d.IsConnected() {
		// panic("DEV NO CONNECT")
		println("Device not connected")
	}
	var x, y, z int32
	for {
		err = d.Update()
		if err != nil {
			println(err.Error())
		}
		x, y, z = d.North()
		println(x, y, z)
		time.Sleep(750 * time.Millisecond)
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
