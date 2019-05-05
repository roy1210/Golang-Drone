package main

import (
	"log"

	"github.com/roy1210/Study/Go-drone/gotello/config"
	"github.com/roy1210/Study/Go-drone/gotello/utils"
)

func main() {
	utils.LoggingSettings(config.Config.LogFile)
	log.Println("test")
	// droneManager := models.NewDroneManager()
	// droneManager.TakeOff()
	// time.Sleep(10 * time.Second)
	// droneManager.Land()
}
