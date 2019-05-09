package models

import (
	"golang.org/x/sync/semaphore"
	"time"

	"gobot.io/x/gobot"
	"gobot.io/x/gobot/platforms/dji/tello"
)

const (
	DefaultSpeed      = 10
	WaitDroneStartSec = 5
)

//3rd partyのファイルを書き換えることはせず、必要な物は自分で足す。
// patrol: Droneが自動で巡回する。
// SemaphoreでパトロールがGoroutineから１つだけ実行するようにする。
type DroneManager struct {
	//tello: robot name. Driverと考えればイメージしやすい。
	*tello.Driver
	Speed int
	patrolSem *semaphore.Weighted
	patrolQuit chan bool
	isPatrolling bool
}

// Droneの基本動作設定
func NewDroneManager() *DroneManager {
	drone := tello.NewDriver("8889")
	droneManager := &DroneManager{
		Driver: drone,
		Speed:  DefaultSpeed,
		patrolSem : semaphore.NewWeighted(1),
		patrolQuit: make(chan bool),
		isPatrolling: false,
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

func (d *DroneManager) Patrol(){
	go func(){
		// 1つだけ、ブロッキングなしでロックを取得できる。
		// Acquire Goroutineで走らせるプログラムの数
		isAcquire := d.patrolSem.TryAcquire(1)
		// ２回目のパトロールでは!isAcquireでロックが取れない。
		// Acquireできるのは１個だけだから。Loopを抜ける
		if !isAcquire{
			d.patrolQuit <- true
			d.isPatrolling = false
			return
		}

		// Loopの最後に１個 Releaseされ、isAcquireでロックを取得できる。
		defer d.patrolSem.Release(1)
		// いまからPatrolする
		// statusの項目を増やしてパトロールの項目を返る。
		// ３秒後にPatrolのStatusを変える。
		d.isPatrolling = true
		status := 0
		t := time.NewTicker(3 * time.Second)

		for{
			select{
				// C: ticker time
			case <-t.C:
				d.Hover()
				switch status{
				case 1:
					d.Forward(d.Speed)
				case 2:
					d.Right(d.Speed)
				case 3:
					d.Backward(d.Speed)
				case 4:
					d.Left(d.Speed)
				case 5:
					status=0
				}
				status++
				// breakの方法.  d.patrolQuit channelがtrueで入って来た場合 
			case <- d.patrolQuit:
				t.Stop()
				d.Hover()
				d.isPatrolling = false
				return

			}
		}
}()
}