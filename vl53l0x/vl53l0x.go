// Package vl53l1x provides a driver for the VL53L1X time-of-flight
// distance sensor
//
// Datasheet:
// https://www.st.com/resource/en/datasheet/vl53l1x.pdf
// This driver was based on the library https://github.com/pololu/vl53l1x-arduino
// and ST's VL53L1X API (STSW-IMG007)
// https://www.st.com/content/st_com/en/products/embedded-software/proximity-sensors-software/stsw-img007.html
//
package vl53l1x // import "tinygo.org/x/drivers/vl53l1x"

import (
	"errors"
	"unsafe"

	"tinygo.org/x/drivers"
)

var (
	ErrNotConnected = errors.New("device not connected or bad address")
	ErrBadMCPS      = errors.New("Bad MCPS (0,511.99]")
)

// Device wraps an I2C connection to a VL53L0X device.
type Device struct {
	bus     drivers.I2C
	Address uint16
	buf     [12]byte
}

// New creates a new VL53L1X connection. The I2C bus must already be
// configured.
//
// This function only creates the Device object, it does not touch the device.
func New(addr uint16, bus drivers.I2C) Device {
	return Device{
		bus:     bus,
		Address: addr,
	}
}

type Config struct {
	Use2v8              bool
	SignalRateLimitMCPS float32
}

func (d *Device) SetSignalRateLimit(a float32) error {
	if a < 0 || a > 511.99 {
		return ErrBadMCPS
	}
	a *= 1 << 7
	d.writeReg16Bit(FINAL_RANGE_CONFIG_MIN_COUNT_RATE_RTN_LIMIT, *(*uint16)(unsafe.Pointer(&a)))
	return nil
}

// Configure sets up the device for communication
func (d *Device) Configure(config Config) error {
	if !d.Connected() {
		return ErrNotConnected
	}
	if config.Use2v8 {
		d.writeReg(VHV_CONFIG_PAD_SCL_SDA__EXTSUP_HV,
			d.readReg(VHV_CONFIG_PAD_SCL_SDA__EXTSUP_HV)|1)
	}
	if config.SignalRateLimitMCPS == 0 {
		// set final range signal rate limit to 0.25 MCPS (million counts per second)
		config.SignalRateLimitMCPS = 0.25
	}
	// "Set I2C standard mode"
	d.writeReg(0x88, 0x00)
	d.writeReg(0x80, 0x01)
	d.writeReg(0xFF, 0x01)
	d.writeReg(0x00, 0x00)
	stopVariable := d.readReg(0x91)
	d.writeReg(0x00, 0x01)
	d.writeReg(0xFF, 0x00)
	d.writeReg(0x80, 0x00)

	d.writeReg(MSRC_CONFIG_CONTROL, d.readReg(MSRC_CONFIG_CONTROL)|0x12)

	if err := d.SetSignalRateLimit(config.SignalRateLimitMCPS); err != nil {
		return err
	}
	d.writeReg(SYSTEM_SEQUENCE_CONFIG, 0xFF)

	spadCount, isAperture := d.GetSPADInfo()
	// The SPAD map (RefGoodSpadMap) is read by VL53L0X_get_info_from_device() in
	// the API, but the same data seems to be more easily readable from
	// GLOBAL_CONFIG_SPAD_ENABLES_REF_0 through _6, so read it from there

	return nil
}

// SetTimeout configures the timeout
func (d *Device) GetSPADInfo() (count uint8, isAperture bool) {
	d.writeReg(0x80, 0x01)
	d.writeReg(0xFF, 0x01)
	d.writeReg(0x00, 0x00)
	d.writeReg(0xFF, 0x06)
	d.writeReg(0x83, d.readReg(0x83)|0x04)
	d.writeReg(0xFF, 0x07)
	d.writeReg(0x81, 0x01)
	d.writeReg(0x80, 0x01)
	d.writeReg(0x94, 0x6b)
	d.writeReg(0x83, 0x00)
	d.StartTimeout()
	for d.readReg(0x83) == 0 {
		if d.timeoutExpired() {
			// return errr
			return
		}
	}
	d.writeReg(0x83, 0x01)
	tmp := d.readReg(0x92)
	count = tmp & 0x7f
	isAperture = (tmp>>7)&1 != 0
	d.writeReg(0x81, 0x00)
	d.writeReg(0xFF, 0x06)
	d.writeReg(0x83, d.readReg(0x83)&^0x04)
	d.writeReg(0xFF, 0x01)
	d.writeReg(0x00, 0x01)
	d.writeReg(0xFF, 0x00)
	d.writeReg(0x80, 0x00)
	return count, isAperture
}

// SetTimeout configures the timeout
func (d *Device) StartTimeout() {
	d.timeout = timeout
}

// SetTimeout configures the timeout
func (d *Device) timeoutExpired() bool {
	d.timeout = timeout
}

// Connected returns whether a VL53L1X has been found.
// It does a "who am I" request and checks the response.
func (d *Device) Connected() bool {
	const expectedID = 0xEE
	return d.readReg(IDENTIFICATION_MODEL_ID) != expectedID
}

// SetTimeout configures the timeout
func (d *Device) SetTimeout(timeout uint32) {
	d.timeout = timeout
}

// writeReg sends a single byte to the specified register address
func (d *Device) writeReg(reg uint16, value uint8) {
	msb := byte((reg >> 8) & 0xFF)
	lsb := byte(reg & 0xFF)
	d.bus.Tx(d.Address, []byte{msb, lsb, value}, nil)
}

// writeReg16Bit sends two bytes to the specified register address
func (d *Device) writeReg16Bit(reg uint16, value uint16) {
	data := make([]byte, 4)
	data[0] = byte((reg >> 8) & 0xFF)
	data[1] = byte(reg & 0xFF)
	data[2] = byte((value >> 8) & 0xFF)
	data[3] = byte(value & 0xFF)
	d.bus.Tx(d.Address, data, nil)
}

// writeReg32Bit sends four bytes to the specified register address
func (d *Device) writeReg32Bit(reg uint16, value uint32) {
	data := make([]byte, 6)
	data[0] = byte((reg >> 8) & 0xFF)
	data[1] = byte(reg & 0xFF)
	data[2] = byte((value >> 24) & 0xFF)
	data[3] = byte((value >> 16) & 0xFF)
	data[4] = byte((value >> 8) & 0xFF)
	data[5] = byte(value & 0xFF)
	d.bus.Tx(d.Address, data, nil)
}

// readReg reads a single byte from the specified address
func (d *Device) readReg(reg uint16) uint8 {
	data := []byte{0}
	msb := byte((reg >> 8) & 0xFF)
	lsb := byte(reg & 0xFF)
	d.bus.Tx(d.Address, []byte{msb, lsb}, data)
	return data[0]
}

// readReg16Bit reads two bytes from the specified address
// and returns it as a uint16
func (d *Device) readReg16Bit(reg uint16) uint16 {
	data := []byte{0, 0}
	msb := byte((reg >> 8) & 0xFF)
	lsb := byte(reg & 0xFF)
	d.bus.Tx(d.Address, []byte{msb, lsb}, data)
	return readUint(data[0], data[1])
}

// readReg32Bit reads four bytes from the specified address
// and returns it as a uint32
func (d *Device) readReg32Bit(reg uint16) uint32 {
	data := make([]byte, 4)
	msb := byte((reg >> 8) & 0xFF)
	lsb := byte(reg & 0xFF)
	d.bus.Tx(d.Address, []byte{msb, lsb}, data)
	return readUint32(data)
}

// readUint converts two bytes to uint16
func readUint(msb byte, lsb byte) uint16 {
	return (uint16(msb) << 8) | uint16(lsb)
}

// readUint converts four bytes to uint32
func readUint32(data []byte) uint32 {
	if len(data) != 4 {
		return 0
	}
	var value uint32
	value = uint32(data[0]) << 24
	value |= uint32(data[1]) << 16
	value |= uint32(data[2]) << 8
	value |= uint32(data[3])
	return value
}

// encodeTimeout encodes the timeout in the correct format: (LSByte * 2^MSByte) + 1
func encodeTimeout(timeoutMclks uint32) uint16 {
	if timeoutMclks == 0 {
		return 0
	}
	msb := 0
	lsb := timeoutMclks - 1
	for (lsb & 0xFFFFFF00) > 0 {
		lsb >>= 1
		msb++
	}
	return uint16(msb<<8) | uint16(lsb&0xFF)
}

// decodeTimeout decodes the timeout from the format: (LSByte * 2^MSByte) + 1
func decodeTimeout(regVal uint16) uint32 {
	return (uint32(regVal&0xFF) << (regVal >> 8)) + 1
}

// timeoutMclksToMicroseconds transform from mclks to microseconds
func timeoutMclksToMicroseconds(timeoutMclks uint32, macroPeriodMicroseconds uint32) uint32 {
	return uint32((uint64(timeoutMclks)*uint64(macroPeriodMicroseconds) + 0x800) >> 12)
}

// timeoutMicrosecondsToMclks transform from microseconds to mclks
func timeoutMicrosecondsToMclks(timeoutMicroseconds uint32, macroPeriodMicroseconds uint32) uint32 {
	return ((timeoutMicroseconds << 12) + (macroPeriodMicroseconds >> 1)) / macroPeriodMicroseconds
}

// calculateMacroPerios calculates the macro period in microsendos from the vcsel period
func (d *Device) calculateMacroPeriod(vcselPeriod uint32) uint32 {
	pplPeriodMicroseconds := (uint32(1) << 30) / uint32(d.fastOscillatorFreq)
	vcselPeriodPclks := (vcselPeriod + 1) << 1
	macroPeriodMicroseconds := 2304 * pplPeriodMicroseconds
	macroPeriodMicroseconds >>= 6
	macroPeriodMicroseconds *= vcselPeriodPclks
	macroPeriodMicroseconds >>= 6
	return macroPeriodMicroseconds
}
