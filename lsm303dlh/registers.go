package lsm303dlh

/*
	Magnetometer and accelerometer register values.
*/

const (
	// According to datasheet I2C Address is 0011110x (x is the I2C read/write bit)
	I2CAddr = 0b0011110
)

// Magnetometer registers
const (
	MAG_CRA_REG_M = 0x00
	MAG_CRB_REG_M = 0x01
	MAG_MR_REG_M  = 0x02
	MAG_OUT_X_H_M = 0x03
	MAG_OUT_X_L_M = 0x04
	MAG_OUT_Z_H_M = 0x05
	MAG_OUT_Z_L_M = 0x06
	MAG_OUT_Y_H_M = 0x07
	MAG_OUT_Y_L_M = 0x08
	MAG_SR_REG_M  = 0x09
	MAG_IRA_REG_M = 0x0A
	MAG_IRB_REG_M = 0x0B
	MAG_IRC_REG_M = 0x0C

	MAG_TEMP_OUT_H_M = 0x31
	MAG_TEMP_OUT_L_M = 0x32
)

const MAG_DEVICE_ID = 0b01000000

// Magnetometer gains
const (
	MAGGAIN_1_3 = 0x20 // +/- 1.3
	MAGGAIN_1_9 = 0x40 // +/- 1.9
	MAGGAIN_2_5 = 0x60 // +/- 2.5
	MAGGAIN_4_0 = 0x80 // +/- 4.0
	MAGGAIN_4_7 = 0xA0 // +/- 4.7
	MAGGAIN_5_6 = 0xC0 // +/- 5.6
	MAGGAIN_8_1 = 0xE0 // +/- 8.1
)

// Magentometer rates
const (
	MAGRATE_0_7 = 0x00 // 0.75 Hz
	MAGRATE_1_5 = 0x01 // 1.5 Hz
	MAGRATE_3_0 = 0x02 // 3.0 Hz
	MAGRATE_7_5 = 0x03 // 7.5 Hz
	MAGRATE_15  = 0x04 // 15 Hz
	MAGRATE_30  = 0x05 // 30 Hz
	MAGRATE_75  = 0x06 // 75 Hz
	MAGRATE_220 = 0x07 // 220 Hz
)
