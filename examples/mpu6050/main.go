// Connects to an MPU6050 I2C accelerometer/gyroscope.
package main

import (
	"machine"
	"time"

	"tinygo.org/x/drivers/mpu6050"
)

func main() {
	machine.I2C0.Configure(machine.I2CConfig{})

	mpuDevice := mpu6050.New(machine.I2C0, mpu6050.DefaultAddress)

	err := mpuDevice.Configure(mpu6050.Config{
		AccelRange: mpu6050.ACCEL_RANGE_16,
		GyroRange:  mpu6050.GYRO_RANGE_2000,
	})
	if err != nil {
		panic(err.Error())
	}
	for {
		time.Sleep(time.Millisecond * 100)
		err := mpuDevice.Update()
		if err != nil {
			println("error reading from mpu6050:", err.Error())
			continue
		}
		print("acceleration: ")
		println(mpuDevice.Acceleration())
		print("angular velocity:")
		println(mpuDevice.AngularVelocity())
		print("temperature celsius:")
		println(mpuDevice.Temperature() / 1000)
	}
}
