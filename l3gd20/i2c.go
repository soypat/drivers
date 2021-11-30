package l3gd20

import (
	"tinygo.org/x/drivers"
)

type DevI2C struct {
	addr uint8
	// sensitivity or range.
	mul int32
	bus drivers.I2C
	buf [1]byte
	// gyro databuf.
	databuf [6]byte
	data    [3]int32
}

func NewI2C(bus drivers.I2C, addr uint8) *DevI2C {
	return &DevI2C{
		addr: addr,
		bus:  bus,
		mul:  sensMul250,
	}
}

// Initializes and configures the device.
func (d *DevI2C) Configure(cfg Config) error {
	err := cfg.validate()
	if err != nil {
		return err
	}

	// Reset then switch to normal mode and enable all three channels.
	err = d.write8(CTRL_REG1, 0)
	if err != nil {
		return err
	}
	err = d.write8(CTRL_REG1, 0x0F)
	if err != nil {
		return err
	}
	// Set sensitivity
	switch cfg.Range {
	case Range_250:
		d.mul = sensMul250
	case Range_500:
		d.mul = sensMul500
	case Range_2000:
		d.mul = sensMul2000
	default:
		return ErrBadRange
	}
	err = d.write8(CTRL_REG4, cfg.Range)
	if err != nil {
		return err
	}
	// Finally verify whomai register and return error if
	// board is not who it says it is. Some counterfeit boards
	// have incorrect whomai but can still be used.
	whoami, err := d.read8(WHOAMI)
	if err != nil {
		return err
	}
	if whoami != expectedWHOAMI && whoami != expectedWHOAMI_H {
		return ErrBadIdentity
	}
	return nil
}

func (d *DevI2C) Update() error {
	err := d.bus.ReadRegister(d.addr, OUT_X_L, d.databuf[:6])
	if err != nil {
		return err
	}
	x := int16(d.databuf[0]) | int16(d.databuf[1])<<8
	y := int16(d.databuf[2]) | int16(d.databuf[3])<<8
	z := int16(d.databuf[4]) | int16(d.databuf[5])<<8
	d.data[0] = d.mul * int32(x)
	d.data[1] = d.mul * int32(y)
	d.data[2] = d.mul * int32(z)
	return nil
}

// returns result in microradians
func (d *DevI2C) AngularVelocity() (x, y, z int32) {
	return d.data[0], d.data[1], d.data[2]
}

// func (d DevI2C) Update(measurement)

func (d DevI2C) read8(reg uint8) (byte, error) {
	err := d.bus.ReadRegister(d.addr, reg, d.buf[:1])
	return d.buf[0], err
}

func (d DevI2C) write8(reg uint8, val byte) error {
	d.buf[0] = val
	return d.bus.WriteRegister(d.addr, reg, d.buf[:1])
}
