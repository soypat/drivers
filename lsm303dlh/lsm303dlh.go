package lsm303dlh

import (
	"encoding/binary"
	"errors"

	"tinygo.org/x/drivers"
)

/*
	Base on https://github.com/adafruit/Adafruit_LSM303DLH_Mag.
*/

const (
	grav              = 9.80665 // Earth's local gravity field at surface.
	gaussToMicrotesla = 100

	gain1_3 = 1000 * gaussToMicrotesla / 980
	gain1_9 = 1000 * gaussToMicrotesla / 760
)

var (
	ErrMagGain      = errors.New("lsm303dlh: bad gain")
	ErrMagRate      = errors.New("lsm303dlh: bad rate")
	ErrNotConnected = errors.New("lsm303dlh: not connected or bad I2C address")
)

// Dev is a LSM303DLH Magnetometer.
type Dev struct {
	addr    uint8
	bus     drivers.I2C
	magGain int32
	// 8 bit read/write buffer
	buf [1]byte
	// magnetometer read buffer
	magbuf [3 * 2]byte
	// Magnetometer values in microradians.
	magVal [3]int32
}

// Config lets user set Gain and Rate values for magnetometer
// when calling Dev.Configure().
type Config struct {
	MagGain uint8
	MagRate uint8
}

func (cfg *Config) validate() error {
	if cfg.MagGain == 0 {
		cfg.MagGain = MAGGAIN_1_3
	}
	if cfg.MagRate > 7 {
		return ErrMagRate
	}
	gb := cfg.MagGain >> 5
	if gb == 0 || gb > 7 {
		return ErrMagGain
	}
	return nil
}

// New instantiates a new  LSM303DLH Magnetometer handle. It performs no IO.
func New(bus drivers.I2C, addr uint8) *Dev {
	return &Dev{
		bus:     bus,
		addr:    addr,
		magGain: 1,
		magVal:  [3]int32{-1, -1, -1},
	}
}

// Configure configures LSM303DLH Magnetometer registers and prepares
// it for reading magnetometer values.
func (d *Dev) Configure(cfg Config) error {
	// Configuration validation
	err := cfg.validate()
	if err != nil {
		return err
	}

	if !d.IsConnected() {
		return ErrNotConnected
	}

	// Configure magnetometer. Start by enabling it.
	err = d.writeMag(MAG_MR_REG_M, 0)
	if err != nil {
		return err
	}

	err = d.writeMag(MAG_CRB_REG_M, cfg.MagGain)
	if err != nil {
		return err
	}
	switch cfg.MagGain {
	case MAGGAIN_1_3:
		d.magGain = 980

	case MAGGAIN_1_9:
		d.magGain = 760

	case MAGGAIN_2_5:
		d.magGain = 600

	case MAGGAIN_4_0:
		d.magGain = 400

	case MAGGAIN_4_7:
		d.magGain = 355

	case MAGGAIN_5_6:
		d.magGain = 295

	case MAGGAIN_8_1:
		d.magGain = 205
	}
	d.magGain = 1000 * gaussToMicrotesla / d.magGain

	// Set Magnetometer rate. Ensures to not modify previous register value.
	craVal, err := d.readMag(MAG_CRA_REG_M)
	if err != nil {
		return err
	}
	craVal &^= 7 << 2
	craVal |= cfg.MagRate << 2
	return d.writeMag(MAG_CRA_REG_M, craVal)
}

// Update reads magnetometer values and stores them in-struct (inside Dev).
// These values can be accessed by calling MagneticField()
func (d *Dev) Update() error {
	// self._read_bytes(self._mag_device, _REG_MAG_OUT_X_H_M, 6, self._BUFFER)
	// raw_values = struct.unpack_from(">hhh", self._BUFFER[0:6])
	// return (raw_values[0], raw_values[2], raw_values[1])
	// Update magnetometer readings
	err := d.bus.ReadRegister(d.addr, MAG_OUT_X_H_M, d.magbuf[:6])
	if err != nil {
		return err
	}
	x := int16(binary.BigEndian.Uint16(d.magbuf[0:]))
	y := int16(binary.BigEndian.Uint16(d.magbuf[2:]))
	z := int16(binary.BigEndian.Uint16(d.magbuf[4:]))
	d.magVal[0] = d.magGain * int32(x)
	d.magVal[1] = d.magGain * int32(y)
	d.magVal[2] = d.magGain * int32(z)
	return nil
}

// MagneticField returns field in nanoteslas.
func (d *Dev) North() (x, y, z int32) {
	// Shift values to create properly formed int16 (high byte first)
	// and then cast the int16 to int32 before multiplying by gain
	return d.magVal[0], d.magVal[1], d.magVal[2]
}

// IsConnected returns true if no error is found
// when writing to I2C bus and if constant register matches expected value.
func (d *Dev) IsConnected() bool {
	// LSM303DLHC has no WHOAMI register, but it has IRx_REG_M that should be constant
	b, err := d.readMag(MAG_IRA_REG_M)
	if err != nil || b != 0b01001000 {
		return false
	}
	return true
}
