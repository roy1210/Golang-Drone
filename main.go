package main

import (
	"time"

	"gobot.io/x/gobot"
	"gobot.io/x/gobot/platforms/dji/tello"
)

func main() {
	drone := tello.NewDriver("8889")

	work := func() {
		drone.TakeOff()

		// afterはTakeOff後の累積のDurationで考える。

		gobot.After(10*time.Second, func() {
			drone.Backward(10)
		})
		gobot.After(15*time.Second, func() {
			drone.Hover()
		})
		gobot.After(20*time.Second, func() {
			drone.Right(10)
		})
		gobot.After(25*time.Second, func() {
			drone.Hover()
		})
		gobot.After(30*time.Second, func() {
			drone.Forward(10)
		})
		gobot.After(35*time.Second, func() {
			drone.Hover()
		})
		gobot.After(40*time.Second, func() {
			drone.Left(10)
		})
		gobot.After(45*time.Second, func() {
			drone.Land()
		})

	}

	robot := gobot.NewRobot(
		"tello",
		[]gobot.Connection{},
		[]gobot.Device{drone},
		work,
	)

	robot.Start()
}
