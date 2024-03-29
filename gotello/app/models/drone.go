package models

import (
	"context"
	"image"
	"image/color"
	"io"
	"io/ioutil"
	"log"
	"math"
	"os/exec"
	"strconv"
	"time"

	"github.com/hybridgroup/mjpeg"
	"gobot.io/x/gobot"
	"gobot.io/x/gobot/platforms/dji/tello"
	"gocv.io/x/gocv"
	"golang.org/x/sync/semaphore"
)

const (
	DefaultSpeed = 10
	// フレーム設定。 重くならないように、小さくする。
	WaitDroneStartSec = 5
	frameX            = 960 / 3
	frameY            = 720 / 3
	// フレームの中間のポイント、座標
	frameCenterX = frameX / 2
	frameCenterY = frameY / 2
	frameArea    = frameX * frameY
	// ffmpeg は3次元の配列を持つ
	frameSize         = frameArea * 3
	faceDetectXMLFile = "./app/models/haarcascade_frontalface_default.xml"
	snapshotsFolder   = "./static/img/snapshots/"
)

// 3rd partyのファイルを書き換えることはせず、必要な物は自分で足す。
// patrol: Droneが自動で巡回する。
// SemaphoreでパトロールがGoroutineから１つだけ実行するようにする。
// ffmpeg: pipe 1 でのストリーミング設定
type DroneManager struct {
	*tello.Driver
	Speed                int
	patrolSem            *semaphore.Weighted
	patrolQuit           chan bool
	isPatrolling         bool
	ffmpegIn             io.WriteCloser
	ffmpegOut            io.ReadCloser
	Stream               *mjpeg.Stream
	faceDetectTrackingOn bool
	isSnapShot           bool
}

// Droneの基本動作設定
func NewDroneManager() *DroneManager {
	drone := tello.NewDriver("8889")

	// ffmpegを走らせる。コマンドを打つ感じで。Pipe 0に書き込む
	// -hwaccel 動画を走らせる時ハードか、ソフトかどっちが良い
	ffmpeg := exec.Command("ffmpeg", "-hwaccel", "auto", "-hwaccel_device", "opencl", "-i", "pipe:0", "-pix_fmt", "bgr24",
		"-s", strconv.Itoa(frameX)+"x"+strconv.Itoa(frameY), "-f", "rawvideo", "pipe:1")

	// 取り込むときはIn,出すときはOut
	ffmpegIn, _ := ffmpeg.StdinPipe()
	ffmpegOut, _ := ffmpeg.StdoutPipe()

	droneManager := &DroneManager{
		Driver:               drone,
		Speed:                DefaultSpeed,
		patrolSem:            semaphore.NewWeighted(1),
		patrolQuit:           make(chan bool),
		isPatrolling:         false,
		ffmpegIn:             ffmpegIn,
		ffmpegOut:            ffmpegOut,
		Stream:               mjpeg.NewStream(),
		faceDetectTrackingOn: false,
		isSnapShot:           false,
	}

	// Gobotのworkパターン
	work := func() {
		if err := ffmpeg.Start(); err != nil {
			log.Println(err)
			return
		}

		// tello.ConnectedEvent : Droneを接続したら何をするか。
		drone.On(tello.ConnectedEvent, func(data interface{}) {
			log.Println("Connected")
			// ビデオをオンにする
			drone.StartVideo()
			drone.SetVideoEncoderRate(tello.VideoBitRateAuto)
			// カメラの露光レベル
			drone.SetExposure(0)

			//　100ミリセカンド毎にビデオのバイナリーを取り続ける。
			gobot.Every(100*time.Millisecond, func() {
				drone.StartVideo()
			})

			// とってきたInをOutでとりたい。Streamにわたす
			droneManager.StreamVideo()
		})

		// drone.OnのVideoFrameが入ってきたときに、ffmpegのInに書き込める
		drone.On(tello.VideoFrameEvent, func(data interface{}) {
			pkt := data.([]byte)
			if _, err := ffmpegIn.Write(pkt); err != nil {
				log.Println(err)
			}
		})
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

func (d *DroneManager) Patrol() {
	go func() {
		// 1つだけ、ブロッキングなしでロックを取得できる。
		// Acquire Goroutineで走らせるプログラムの数
		isAcquire := d.patrolSem.TryAcquire(1)
		// ２回目のパトロールでは!isAcquireでロックが取れない。
		// Acquireできるのは１個だけだから。Loopを抜ける
		if !isAcquire {
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
		for {
			select {
			// C: ticker time
			case <-t.C:
				d.Hover()
				switch status {
				case 1:
					d.Forward(d.Speed)
				case 2:
					d.Right(d.Speed)
				case 3:
					d.Backward(d.Speed)
				case 4:
					d.Left(d.Speed)
				case 5:
					status = 0
				}
				status++
				// breakの方法.  d.patrolQuit channelがtrueで入って来た場合
			case <-d.patrolQuit:
				t.Stop()
				d.Hover()
				d.isPatrolling = false
				return
			}
		}
	}()
}

func (d *DroneManager) StartPatrol() {
	// 0 valueは False
	if !d.isPatrolling {
		d.Patrol()
	}
}

func (d *DroneManager) StopPatrol() {

	if d.isPatrolling {
		d.Patrol()
	}
}

func (d *DroneManager) StreamVideo() {
	go func(d *DroneManager) {
		classifier := gocv.NewCascadeClassifier()
		defer classifier.Close()
		// 実行もされているイメージ
		if !classifier.Load(faceDetectXMLFile) {
			log.Printf("Error reading cascade file: %v\n", faceDetectXMLFile)
			return
		}

		blue := color.RGBA{0, 0, 255, 0}

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

			if d.faceDetectTrackingOn {
				d.StopPatrol()
				rects := classifier.DetectMultiScale(img)
				log.Printf("found %d faces\n", len(rects))
				// index は省く

				if len(rects) == 0 {
					d.Hover()
				}
				for _, r := range rects {
					gocv.Rectangle(&img, r, blue, 3)
					// Humanの位置
					pt := image.Pt(r.Max.X, r.Min.Y-5)
					gocv.PutText(&img, "Human", pt, gocv.FontHersheyPlain, 1.2, blue, 2)

					faceWidth := r.Max.X - r.Min.X
					faceHight := r.Max.Y - r.Min.Y
					// X 20 50 => 20 + (30/2) = 35
					faceCenterX := r.Min.X + (faceWidth / 2)
					faceCenterY := r.Min.Y + (faceHight / 2)
					faceArea := faceWidth * faceHight
					// 160 - 35 = 125
					diffX := frameCenterX - faceCenterX
					diffY := frameCenterY - faceCenterY
					percentF := math.Round(float64(faceArea) / float64(frameArea) * 100)

					move := false
					if diffX < -20 {
						d.Right(15)
						move = true
					}
					if diffX > 20 {
						d.Left(15)
						move = true
					}

					if diffY < -30 {
						d.Down(25)
						move = true
					}

					if diffY > 30 {
						d.Up(25)
						move = true
					}
					if percentF > 7.0 {
						d.Backward(10)
						move = true
					}
					if percentF < 0.9 {
						d.Forward(10)
						move = true
					}
					if !move {
						d.Hover()
					}
					break
				}
			}

			jpegBuf, _ := gocv.IMEncode(".jpg", img)

			if d.isSnapShot {
				backupFileName := snapshotsFolder + time.Now().Format(time.RFC3339) + ".jpg"
				ioutil.WriteFile(backupFileName, jpegBuf, 0644)

				// snapshot.jpgは上書きされるから、↑で保存する。
				snapshotFileName := snapshotsFolder + "snapshot.jpg"
				ioutil.WriteFile(snapshotFileName, jpegBuf, 0644)
				d.isSnapShot = false
			}

			d.Stream.UpdateJPEG(jpegBuf)
		}
	}(d)
}

// contextを使用し、キャンセル条件を書く
func (d *DroneManager) TakeSnapShot() {
	d.isSnapShot = true
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	for{
		if !d.isSnapShot || ctx.Err() !=nil{
			break
		}
	}
	d.isSnapShot = false
}

func (d *DroneManager) EnableFaceDetectTracking() {
	d.faceDetectTrackingOn = true
}

func (d *DroneManager) DisableFaceDetectTracking() {
	d.faceDetectTrackingOn = false
	d.Hover()
}
