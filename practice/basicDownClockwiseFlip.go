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
			drone.Down(20)
		})
		gobot.After(15*time.Second, func() {
			drone.Hover()
		})
		gobot.After(20*time.Second, func() {
			drone.Up(20)
		})
		gobot.After(25*time.Second, func() {
			drone.Hover()
		})
		gobot.After(30*time.Second, func() {
			drone.Clockwise(50)
		})
		gobot.After(35*time.Second, func() {
			drone.Hover()
		})
		gobot.After(40*time.Second, func() {
			drone.CounterClockwise(50)
		})
		gobot.After(45*time.Second, func() {
			drone.Hover()
		})
		gobot.After(50*time.Second, func() {
			drone.FrontFlip()
		})
		gobot.After(55*time.Second, func() {
			drone.Hover()
		})
		gobot.After(60*time.Second, func() {
			drone.BackFlip()
		})
		gobot.After(65*time.Second, func() {
			drone.Hover()
		})
		gobot.After(70*time.Second, func() {
			drone.RightFlip()
		})
		gobot.After(75*time.Second, func() {
			drone.Hover()
		})
		gobot.After(80*time.Second, func() {
			drone.LeftFlip()
		})
		gobot.After(85*time.Second, func() {
			drone.Hover()
		})
		gobot.After(90*time.Second, func() {
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
