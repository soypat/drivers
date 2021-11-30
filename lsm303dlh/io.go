package lsm303dlh

func (d Dev) write(addr, reg uint8, val byte) error {
	d.buf[0] = val
	return d.bus.WriteRegister(addr, reg, d.buf[:1])
}

func (d Dev) read(addr, reg uint8) (byte, error) {
	d.buf[0] = 0
	err := d.bus.WriteRegister(addr, reg, d.buf[:1])
	return d.buf[0], err
}

func (d Dev) writeMag(reg uint8, val byte) error {
	return d.write(d.addr, reg, val)
}
func (d Dev) readMag(reg uint8) (byte, error) {
	return d.read(d.addr, reg)
}
