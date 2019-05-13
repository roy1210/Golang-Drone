package main

import (
	"io"
	"log"
	"os/exec"
	"strconv"

	"github.com/hybridgroup/mjpeg"
	"gobot.io/x/gobot/platforms/dji/tello"
	"gocv.io/x/gocv"
	"golang.org/x/sync/semaphore"
)

const (
	DefaultSpeed      = 10
	WaitDroneStartSec = 5
	// フレーム設定。 重くならないように、小さくする。
	frameX = 960 / 3
	frameY = 720 / 3
	// フレームの中間のポイント、座標
	frameCenterX = frameX / 2
	frameCenterY = frameY / 2
	frameArea    = frameX * frameY
	// ffmpeg は3次元の配列を持つ
	frameSize = frameArea * 3
)

// 3rd partyのファイルを書き換えることはせず、必要な物は自分で足す。
// patrol: Droneが自動で巡回する。
// SemaphoreでパトロールがGoroutineから１つだけ実行するようにする。
type DroneManager struct {
	//tello: robot name. Driverと考えればイメージしやすい。
	*tello.Driver
	Speed        int
	patrolSem    *semaphore.Weighted
	patrolQuit   chan bool
	isPatrolling bool
	// pipe 1 でのストリーミング設定
	ffmpegIn  io.WriteCloser
	ffmpegOut io.ReadCloser
	Stream    *mjpeg.Stream
}

// Droneの基本動作設定
func NewDroneManager() *DroneManager {
	drone := tello.NewDriver("8889")

	// ffmpegを走らせる。コマンドを打つ感じで。Pipe 0に書き込む
	// -hwaccel 動画を走らせる時ハードか、ソフトかどっちが良い
	ffmpeg := exec.Command("ffmpeg", "-hwaccel", "auto", "-hwaccel_device", "opencl", "-i", "pipe:0", "-pix_fmt", "bgr24", "-s", strconv.Itoa(frameX)+"x"+strconv.Itoa(frameY), "-f", "rawvideo", "pipe:1")

	// 取り込むときはIn,出すときはOut
	ffmpegIn, _ := ffmpeg.StdinPipe()
	ffmpegOut, _ := ffmpeg.StdoutPipe()

	droneManager := &DroneManager{
		Driver:       drone,
		Speed:        DefaultSpeed,
		patrolSem:    semaphore.NewWeighted(1),
		patrolQuit:   make(chan bool),
		isPatrolling: false,
		ffmpegIn:     ffmpegIn,
		ffmpegOut:    ffmpegOut,
		Stream:       mjpeg.NewStream(),
	}

	work := func() {
		if err := ffmpeg.Start(); err != nil {
			log.Println(err)
			return
		}
		return droneManager
	}
}

func (d *DroneManager) StartPatrol() {
	// 0 valueは False
	if !d.isPatrolling {
		d.Patrol()
	}
}

func (d *DroneManager) StopPatrol() {
	// Semaphoreは2回目のアクセスでFalseとなる。
	if d.isPatrolling {
		d.Patrol()
	}
}

func (d *DroneManager) StreamVideo() {
	go func(d *DroneManager) {
		for {
			buf := make([]byte, frameSize)
			if _, err := io.ReadFull(d.ffmpegOut, buf); err != nil {
				log.Println(err)
			}
			// とってきたByte配列をImageに変換
			img, _ := gocv.NewMatFromBytes(frameY, frameX, gocv.MatTypeCV8UC3, buf)
			if img.Empty() {
				continue
			}

			jpegBuf, _ := gocv.IMEncode(".jpg", img)
			d.Stream.UpdateJPEG(jpegBuf)
		}
	}(d)
}

func main() {
	//todo
}
