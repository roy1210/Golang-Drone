package main

import (
	"log"

	"github.com/roy1210/Study/Go-drone/gotello/app/controllers"
	"github.com/roy1210/Study/Go-drone/gotello/config"
	"github.com/roy1210/Study/Go-drone/gotello/utils"
)

func main() {
	utils.LoggingSettings(config.Config.LogFile)
	log.Println(controllers.StartWebServer())
}

// func main() {
// 	utils.LoggingSettings(config.Config.LogFile)
// 	log.Println("test")
// 	droneManager := models.NewDroneManager()
// 	droneManager.TakeOff()
// 	time.Sleep(10 * time.Second)
// 	droneManager.Patrol()
// 	time.Sleep(30 * time.Second)
// 	// 2回目のPatrolはロックが取れないから、Loopを終了させる.
// 	droneManager.Patrol()
// 	// gobot.Afterの秒数換算と間違えないよう注意
// 	time.Sleep(10 * time.Second)
// 	droneManager.Land()
// }
