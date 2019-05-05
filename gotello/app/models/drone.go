package models

import (
	"time"

	"gobot.io/x/gobot"
	"gobot.io/x/gobot/platforms/dji/tello"
)

const (
	DefaultSpeed      = 10
	WaitDroneStartSec = 5
)

//3rd partyのファイルを書き換えることはせず、必要な物は自分で足す。
type DroneManager struct {
	//tello: robot name. Driverと考えればイメージしやすい。
	*tello.Driver
	Speed int
}

// Droneの基本動作設定
func NewDroneManager() *DroneManager {
	drone := tello.NewDriver("8889")
	droneManager := &DroneManager{
		Driver: drone,
		Speed:  DefaultSpeed,
	}
	work := func() {
		//todo
	}
	robot := gobot.NewRobot("tello", []gobot.Connection{}, []gobot.Device{drone}, work)
	// 他のコードも走らせられるよう、Goroutineにする
	go robot.Start()
	// Goroutineの後は、インターバルを入れるイメージ。memoryのinvalid errorを防ぐため。
	time.Sleep(WaitDroneStartSec * time.Second)

	// mainにdroneManager つまりembeddedされたDroneManager => Driverが返る。
	// droneManager.Left()とかで操作できるようにしたい
	return droneManager
}
