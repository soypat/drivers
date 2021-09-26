package drivers

// Measurement specifies a type of measurement,
// for example: temperature, acceleration, pressure.
type Measurement uint32

// Sensor measurements
const (
	Temperature Measurement = 1 << iota
	Humidity
	Pressure
	Distance
	Acceleration
	AngularVelocity
	MagneticField
	Luminosity
	Time
	Audio
)

// Sensor represents an object capable of making one
// or more measurements. A sensor will then have methods
// which read the last updated measurements.
//
// Many Sensors may be collected into
// one Sensor interface to synchronize measurements.
type Sensor interface {
	Update(which Measurement) error
}
