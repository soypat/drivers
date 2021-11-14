package enc28j60

import (
	"runtime/interrupt"
	"time"
)

// read len(data) bytes from buffer
func (d *Dev) readBuffer(data []byte) {
	d.enableCS()
	d.bus.Transfer(READ_BUF_MEM)
	d.bus.Tx(nil, data)
	d.disableCS()
	dbp("read from ebuff", data)
}

// write data to TX buffer
func (d *Dev) writeBuffer(data []byte) {
	d.enableCS()
	d.bus.Transfer(WRITE_BUF_MEM)
	d.bus.Tx(data, nil)
	d.disableCS()
	dbp("write to ebuff", data)
}

// the ENC28J60 has 4 banks (0 through 3). It can only read/write to
// one at a time, and much switch between them by writing to ECON1 register.
func (d *Dev) setBank(address uint8) {
	bank := address & BANK_MASK
	if bank != d.Bank {
		d.writeOp(BIT_FIELD_CLR, ECON1, ECON1_BSEL1|ECON1_BSEL0)
		d.writeOp(BIT_FIELD_SET, ECON1, bank>>5)
		d.Bank = bank
	}
}

// readOp reads from a register defined in registers.go. It requires
// the ENC28J60 Bank be set beforehand.
func (d *Dev) readOp(op, address uint8) uint8 {
	d.enableCS()
	d.bus.Transfer(op | (address & ADDR_MASK))
	read, _ := d.bus.Transfer(0)
	// do dummy read if needed (for mac and mii, see datasheet page 29)
	if address&SPRD_MASK != 0 {
		d.bus.Transfer(0)
	}
	d.disableCS()
	return read
}

// readOp writes to a register defined in registers.go. It requires
// the ENC28J60 Bank be set beforehand.
func (d *Dev) writeOp(op, address, data uint8) {
	d.enableCS()
	d.bus.Transfer(op | (address & ADDR_MASK))
	_, err := d.bus.Transfer(data)
	d.disableCS()
	if err != nil {
		dbp("writeOp", d.buf[:1])
	}
}

func (d *Dev) read(address uint8) uint8 {
	d.setBank(address)
	return d.readOp(READ_CTL_REG, address)
}

func (d *Dev) write(address, data uint8) {
	d.setBank(address)
	d.writeOp(WRITE_CTL_REG, address, data)
}

// write16 writes to two contiguous 8 bit addresses (LSB first).
func (d *Dev) write16(addressL uint8, data uint16) {
	d.setBank(addressL)
	d.writeOp(WRITE_CTL_REG, addressL, uint8(data))
	d.writeOp(WRITE_CTL_REG, addressL+1, uint8(data>>8))
}

// write16 reads two contiguous 8 bit addresses and returns
// 16bit value LSB first.
func (d *Dev) read16(addressL uint8) uint16 {
	d.setBank(addressL)
	return uint16(d.readOp(READ_CTL_REG, addressL)) + uint16(d.readOp(READ_CTL_REG, addressL+1))<<8
}

func (d *Dev) phyWrite(address uint8, data uint16) {
	// set the PHY register address. this begins the transaction
	d.write(MIREGADR, address)
	// write the PHY data
	d.write16(MIWRL, data)
	// wait until the PHY write completes
	d.waitOnMISTAT()
}

func (d *Dev) phyRead(address uint8) uint16 {
	// set the PHY register address
	d.write(MIREGADR, address)
	d.writeOp(BIT_FIELD_SET, MICMD, MICMD_MIIRD)
	// Poll the MISTAT.BUSY bit to be
	// certain that the operation is complete.
	if d.waitOnMISTAT() != nil {
		return 0
	}
	// set bank 2 again
	d.setBank(MICMD)
	d.writeOp(BIT_FIELD_CLR, MICMD, MICMD_MIIRD)
	// write the PHY data
	return d.read16(MIRDL)
}

func (d *Dev) waitOnMISTAT() error {
	stat := d.read(MISTAT)
	for stat&MISTAT_BUSY != 0 {
		time.Sleep(time.Microsecond * 15)
		stat = d.read(MISTAT)
		if stat == 0xff { // if read bits are all 1, then there's probably a connection issue
			return ErrIO
		}
	}
	return nil
}

// enableCS enables SPI communication on bus. Disables Interrupts.
// do not call enableCS twice before calling disable
//go:inline
func (d *Dev) enableCS() {
	d.is = interrupt.Disable()
	d.CSB.Low()
}

// disableCS ends SPI communication on bus
// always call disableCS after calling enable once
// critical part done
//go:inline
func (d *Dev) disableCS() {
	d.CSB.High()
	interrupt.Restore(d.is)
}
