package main

import (
	"bytes"
	"encoding/base64"
	"flag"
	"fmt"
	htemplate "html/template"
	"image"
	"image/color"
	"image/draw"
	"image/jpeg"
	"log"
	"math/rand"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
	ttemplate "text/template"
	"time"

	rice "github.com/GeertJohan/go.rice"
	"github.com/go-vgo/robotgo"
	"github.com/gorilla/websocket"
	"github.com/kbinani/screenshot"
)


func main() {
	log.SetFlags(0)
	http.HandleFunc("/script", script)
	http.HandleFunc("/style", style)
	http.HandleFunc("/", home)
	fmt.Println("IP Addresses:")
	host, _ := os.Hostname()
	addrs, _ := net.LookupIP(host)
	for _, addr := range addrs {
		if ipv4 := addr.To4(); ipv4 != nil {
			fmt.Println("IPv4: ", ipv4)
		}
	}
	password = "Generated"
        passwordPtr := flag.String("pass", password, "the desired password, will generate one by default")
        fpsPtr := flag.Int("fps", 60, "the framerate at which the app will start")
	portPtr := flag.Int("port", 80, "the port that te app will be hosted on")
	sessionsPtr := flag.Int("sessions", 5, "the maximum number of websocket endpoints to be created")
	flag.Parse()
	password = *passwordPtr
	fps = *fpsPtr
	port := *portPtr
	sessions = *sessionsPtr
	if strings.Compare(password, "Generated") == 0 {
		password = randSeq(6)
	}
	fmt.Println("Password: " + password)
	fmt.Println("StartingFPS: " + strconv.Itoa(fps))
	fmt.Println("Port: " + strconv.Itoa(port))
	fmt.Println("Sessions: " + strconv.Itoa(sessions))
	calculateFrameTime()
	go makeImage()
	log.Fatal(http.ListenAndServe("0.0.0.0:" + strconv.Itoa(port), nil))
}

func makeImage() {
	for {
		time.Sleep(frameTime)
		if !active {
			return
		}
		imageStart := time.Now()
		all := screenshot.GetDisplayBounds(0).Union(image.Rect(0, 0, 0, 0))
		img, err := screenshot.Capture(all.Min.X, all.Min.Y, all.Dx(), all.Dy())
		if err != nil {
			fmt.Println("Screenshot: ", err)
		}
		x, y := robotgo.GetMousePos()
		c := color.White
		r, g, b, a := img.At(x, y).RGBA()
		_ = a
		if r > 40000 && g > 40000 && b > 40000 {
			c = color.Black
		}
		draw.Draw(img, image.Rect(x-5, y-5, x+5, y+5), &image.Uniform{c}, image.ZP, draw.Src)
		var buff bytes.Buffer
		jpeg.Encode(&buff, img, &jpeg.Options{Quality: 50})
		lastScreen = "data:image/jpeg;base64," + base64.StdEncoding.EncodeToString(buff.Bytes())
		checkReduceFPS(imageStart)
	}
}

func authenticate(w http.ResponseWriter, r *http.Request) bool {
	query := r.URL.Query()
	pass := query.Get("password")
	if pass != password {
		w.WriteHeader(403)
		return false
	}
	return true
}

func input(w http.ResponseWriter, r *http.Request) {
	if !authenticate(w, r) {
		return
	}
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Print("upgrade:", err)
		return
	}
	for {
		mt, message, err := c.ReadMessage()
		if err != nil {
			fmt.Println("read:", err)
			break
		}
		_ = mt
		switch string(string(message)[0]) {
		case "M":
			switch string(string(message)[2]) {
			case "M":
				switch string(string(message)[4]) {
				case "U":
					robotgo.MoveSmoothRelative(0, -20, 3.0, float64(20))
				case "D":
					robotgo.MoveSmoothRelative(0, 20, 3.0, float64(20))
				case "L":
					robotgo.MoveSmoothRelative(-20, 0, 3.0, float64(20))
				case "R":
					robotgo.MoveSmoothRelative(20, 0, 3.0, float64(20))
				}
			case "C":
				switch string(string(message)[4]) {
				case "L":
					robotgo.MouseClick("left", false)
				case "R":
					robotgo.MouseClick("right", false)
				}
			case "A":
				x, errX := strconv.Atoi(strings.Split(string(message), "-")[2])
				y, errY := strconv.Atoi(strings.Split(string(message), "-")[3])
				if errX == nil && errY == nil {
					robotgo.Move(x, y)
				}
			}
		case "K":
			keyCode := ""
			if len(string(message)) > 3 {
				keyCode = strings.Replace(strings.Replace(string(strings.ToLower(string(message)[4:len(string(message))])), "key", "", -1), "arrow", "", -1)
			}
			switch string(string(message)[2]) {
			case "T":
				robotgo.KeyTap(keyCode)
			case "Q":
				keyBuffer = append(keyBuffer, keyCode)
			case "W":
				if len(keyBuffer) > 1 {
					robotgo.KeyTap(keyBuffer[0], keyBuffer[1:len(keyBuffer)])
				}
				keyBuffer = nil
			case "E":
				keyBuffer = nil
			}

		}
	}
}

func screen(w http.ResponseWriter, r *http.Request) {
	if !authenticate(w, r) {
		return
	}
	if checkSocketActive(w,r) {
		return
	}
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Print("upgrade:", err)
		return
	}
	for {
		mt, message, err := c.ReadMessage()
		if err != nil {
			fmt.Println("read:", err)
			deactivateSocket(strings.Split(r.URL.Path, "_")[1])
			determineGlobalActivity()
			break
		}
		if string(message) == "go" {
			activateSocket(strings.Split(r.URL.Path, "_")[1])
			determineGlobalActivity()
			go socketActivity(c, mt, strings.Split(r.URL.Path, "_")[1])
		}
		if string(message) == "stop" {
			c.Close()
			deactivateSocket(strings.Split(r.URL.Path, "_")[1])
			determineGlobalActivity()
			break
		}
	}
}

func socketActivity(c *websocket.Conn, mt int, id string) {
	for {
		time.Sleep(frameTime)
		if lastScreen != "" {
			err := c.WriteMessage(mt, []byte(lastScreen))
			if err != nil {
				fmt.Println("write:", err)
				deactivateSocket(id)
				determineGlobalActivity()
				break
			}
		}
		if !active {
			break
		}
	}
}
func home(w http.ResponseWriter, r *http.Request) {
	if checkAvailableSockets() {
		fmt.Println("no more sessions!")
		w.WriteHeader(503)
	} else {
		homeTemplate.Execute(w, nil)
	}
}

func script(w http.ResponseWriter, r *http.Request) {
	socket := strconv.Itoa(determineSocket())
	sockets := map[string]interface{}{
		"screen": "ws://" + r.Host + "/screen_" + socket,
		"input":  "ws://" + r.Host + "/input_" + socket,
	}
	scriptTemplate.Execute(w, sockets)
}

func style(w http.ResponseWriter, r *http.Request) {
	styleTemplate.Execute(w, nil)
}

func randSeq(n int) string {
	rand.Seed(time.Now().Unix())
	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

func calculateFrameTime() {
	frameTime = time.Duration(1000/fps) * time.Millisecond
}

func checkReduceFPS(start time.Time) {
	if time.Now().UnixMilli()-start.UnixMilli() > frameTime.Milliseconds() {
		fps = fps / 2
	} else {
		fps = fps * 2
	}
	fmt.Printf("FPS: " + strconv.Itoa(fps))
	fmt.Printf("\r")
	calculateFrameTime()
}

func determineSocket () int {
	for i := range sockets {
		if !sockets[i] {
			fmt.Println("assign socket: " + strconv.Itoa(i))
			return i
		}
	}
	sockets = append(sockets, false)
	fmt.Println("create and assign socket: " +  strconv.Itoa(len(sockets) - 1))
	http.HandleFunc("/screen_" + strconv.Itoa(len(sockets) - 1), screen)
	http.HandleFunc("/input_" +  strconv.Itoa(len(sockets) - 1), input)
	return len(sockets) - 1
}

func activateSocket (id string) {
	fmt.Println("activate socket: " + id)
	intId, err := strconv.Atoi(id)
	if err == nil {
		sockets[intId] = true
	}
}

func deactivateSocket (id string) {
	fmt.Println("deactivate socket: " + id)
	intId, err := strconv.Atoi(id)
	if err == nil {
		sockets[intId] = false
	}
}

func checkSocketActive (w http.ResponseWriter, r *http.Request) bool {
	id, err := strconv.Atoi(strings.Split(r.URL.Path, "_")[1])
	if err == nil && len(sockets) > id {
		if sockets[id] {
			fmt.Println("socket already active: " + strconv.Itoa(id))
			w.WriteHeader(503)
			return true
		}
	}
	return false
}

func checkAvailableSockets () bool {
	activeSockets := 0
	for i := range sockets {
		if sockets[i] {
			activeSockets++
		}
		i++
	}
	return activeSockets == sessions
}


func determineGlobalActivity () {
	for i := range sockets {
		if sockets[i] && !active{
			active = true
			go makeImage()
			fmt.Println("activate render")
			return
		}
	}
	fmt.Println("deactivate render")
	active = false
}

var box = rice.MustFindBox("frontend")
var tmpl, errTMPL = box.String("index.tmpl")
var css, errCSS = box.String("style.css")
var js, errJS = box.String("script.js")
var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890")
var homeTemplate, errH = htemplate.New("home").Parse(tmpl)
var scriptTemplate, errS = ttemplate.New("script").Parse(js)
var styleTemplate, errC = ttemplate.New("style").Parse(css)
var upgrader = websocket.Upgrader{}
var keyBuffer []string
var password string
var fps int
var sessions int
var frameTime time.Duration
var lastScreen = ""
var active = false
var sockets []bool
