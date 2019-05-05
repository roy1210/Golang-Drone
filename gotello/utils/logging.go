package utils

import (
	"io"
	"log"
	"os"
)

func LoggingSettings(logFile string) {
	logfile, err := os.OpenFile(logFile, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("file=logFile err=%s", err.Error())
	}

	//os.Stdout (エラー画面) とlogfileの2手に書きたいのでmultiLogFile
	multiLogFile := io.MultiWriter(os.Stdout, logfile)
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)

	log.SetOutput(multiLogFile)
}
