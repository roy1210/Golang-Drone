package controllers

import (
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"regexp"
	"strconv"

	"github.com/roy1210/Study/Go-drone/gotello/app/models"
	"github.com/roy1210/Study/Go-drone/gotello/config"
)

// 引数のテンプレートを描画する
func getTemplate(temp string) (*template.Template, error) {
	return template.ParseFiles("app/views/layout.html", temp)
}

func viewIndexHandler(w http.ResponseWriter, r *http.Request) {
	// t, _ := template.ParseFiles("app/views/index.html")
	t, _ := getTemplate("app/views/index.html")

	err := t.Execute(w, nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func viewControllerHandler(w http.ResponseWriter, r *http.Request) {
	t, _ := getTemplate("app/views/controller.html")

	err := t.Execute(w, nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

var appContext struct {
	DroneManager *models.DroneManager
}

func init() {
	appContext.DroneManager = models.NewDroneManager()
}

func getSpeed(r *http.Request) int {
	strSpeed := r.FormValue("speed")
	if strSpeed == "" {
		return models.DefaultSpeed
	}
	speed, err := strconv.Atoi(strSpeed)
	if err != nil {
		return models.DefaultSpeed
	}
	return speed
}

type APIResult struct {
	// structやstringを渡したいから、万能Typeのinterface{} typeとする
	Result interface{} `json:"result"`
	Code   int         `json:"code"`
}

// Marshalでjsonにする
func APIResponse(w http.ResponseWriter, result interface{}, code int) {
	res := APIResult{Result: result, Code: code}
	js, err := json.Marshal(res)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	// httpのHeaderに「Jsonを返しますよ」の記載をする。
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(js)
}

var apiValidPath = regexp.MustCompile("^/api/(command|shake|video)")

// 先にRegexでの判定を走らせたいから、このFuncを先に走って、後にapiCommandHandlerを走らせる。Wrapする形で。
func apiMakeHandler(fn func(w http.ResponseWriter, r *http.Request)) http.HandlerFunc {
	// fnを実行する前にRegexの確認
	return func(w http.ResponseWriter, r *http.Request) {
		m := apiValidPath.FindStringSubmatch(r.URL.Path)
		// FindStringSubmatchはスライスを返す。エラーならnilになる。（空のスライス）
		if len(m) == 0 {
			// フロントからバックへはJsonで返ってくるので、このエラー処理もJsonで返す。
			APIResponse(w, "Not found", http.StatusNotFound)
		}
		// Regecが通ったら、fnをが走る。これはapiCommandHandlerが入る。
		fn(w, r)
	}
}

// frontからcommandの値を受け取る。
// switch文を使いdroneにcommandの値を渡す
func apiCommandHandler(w http.ResponseWriter, r *http.Request) {

	command := r.FormValue("command")
	log.Printf("action=apiCommandHandler command=%s", command)

	drone := appContext.DroneManager
	switch command {
	case "caseRotation":
		drone.CeaseRotation()
	case "takeOff":
		drone.TakeOff()
	case "land":
		drone.Land()
	case "hover":
		drone.Hover()
	case "up":
		drone.Up(drone.Speed)
	case "clockwise":
		drone.Clockwise(drone.Speed)
	case "counterClockwise":
		drone.CounterClockwise(drone.Speed)
	case "down":
		drone.Down(drone.Speed)
	case "forward":
		drone.Forward(drone.Speed)
	case "left":
		drone.Left(drone.Speed)
	case "right":
		drone.Right(drone.Speed)
	case "backward":
		drone.Backward(drone.Speed)
	case "frontFlip":
		drone.FrontFlip()
	case "leftFlip":
		drone.LeftFlip()
	case "rightFlip":
		drone.RightFlip()
	case "backFlip":
		drone.BackFlip()
	case "patrol":
		drone.Patrol()
	case "stopPatrol":
		drone.StopPatrol()
	// 5秒以内に投げる
	case "throwTakeOff":
		drone.ThrowTakeOff()
	// ドローンが上下左右に動く動く
	case "bounce":
		drone.Bounce()
	case "speed":
		drone.Speed = getSpeed(r)
	default:
		APIResponse(w, "Not found", http.StatusNotFound)
		return
	}

	APIResponse(w, "OK", http.StatusOK)
}

// 実際に返ってきたlog： 2019/05/09 17:03:26 webserver.go:78: action=apiCommandHandler command=ceaseRoatation

func StartWebServer() error {
	http.HandleFunc("/", viewIndexHandler)
	http.HandleFunc("/controller/", viewControllerHandler)
	http.HandleFunc("/api/command/", apiMakeHandler(apiCommandHandler))
	http.Handle("/video/streaming/", appContext.DroneManager.Stream)

	// staticのサーバー立ち上げ。
	// Handlerではなく、既にフォルダとして静的なサイトの準備ができたものに対し、フォルダを読み込んでサーバーからアクセス出来るようにする。CSSやImgの格納場所
	// http.StripPrefix("/static/" : staticがURLの先頭に来たときに"static"フォルダから読む。
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	return http.ListenAndServe(fmt.Sprintf("%s:%d", config.Config.Address, config.Config.Port), nil)
}
