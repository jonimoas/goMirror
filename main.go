package main

import (
	"bytes"
	"context"
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
	portPtr := flag.Int("port", 80, "the port that the app will be hosted on")
	sessionsPtr := flag.Int("sessions", 5, "the maximum number of websocket endpoints to be created")
	speedPtr := flag.Bool("maxspeed", false, "if enabled, no threads sleep and endpoints are not cleaned up. Might make app and system unstable. :)")
	compressionPtr := flag.Int("compression", 50, "the amount of compression that will be applied on the images")
	flag.Parse()
	password = *passwordPtr
	fps = *fpsPtr
	port = *portPtr
	sessions = *sessionsPtr
	speed = *speedPtr
	compression = *compressionPtr
	if strings.Compare(password, "Generated") == 0 {
		password = randSeq(6)
	}
	fmt.Println("Password: " + password)
	fmt.Println("StartingFPS: " + strconv.Itoa(fps))
	fmt.Println("Port: " + strconv.Itoa(port))
	fmt.Println("Sessions: " + strconv.Itoa(sessions))
	fmt.Println("Compression: " + strconv.Itoa(compression))
	calculateFrameTime()
	if (speed) {
		fmt.Println("RUNNING AT MAX SPEED")
	} else {
		go showStatus()
	}
	maintainServer()
}

func showStatus() {
	for {
		time.Sleep(frameTime)
		activeSockets := 0
		for i := range sockets {
			if sockets[i] {
				activeSockets++
			}
			i++
		}
		fmt.Printf("\r")
		fmt.Printf("                                                                       ")
		fmt.Printf("FPS: " + strconv.Itoa(fps) + ", sockets: " + strconv.Itoa(activeSockets) + "/" + strconv.Itoa(len(sockets)) + "/" + strconv.Itoa(sessions) + ", renderer: " + strconv.FormatBool(active))
		fmt.Printf("\r")
	}
}

func maintainServer() {
	for {
		time.Sleep(frameTime)
		if !server {
			setupServer()
		}
	}
}

func setupServer() {
	fmt.Println("Server CleanUp")
	sockets = nil
	mobileSockets = nil
	m = http.NewServeMux()
	m.HandleFunc("/script", script)
	m.HandleFunc("/style", style)
	m.HandleFunc("/", home)
	s = http.Server{Addr: "0.0.0.0:" + strconv.Itoa(port), Handler: m}
	server = true
	fmt.Println("Server Started")
	err := s.ListenAndServe()
	if err != nil {
		server = false
	}
}

func makeImage() {
	for {
		if !active {
			return
		}
		imageStart := time.Now()
		all := screenshot.GetDisplayBounds(0).Union(image.Rect(0, 0, 0, 0))
		img, err := screenshot.Capture(all.Min.X, all.Min.Y, all.Dx(), all.Dy())
		if err != nil {
			fmt.Println("Screenshot: ", err)
		}
		if mobileMode {
			x, y := robotgo.Location()
			c := color.White
			r, g, b, a := img.At(x, y).RGBA()
			_ = a
			if r > 40000 && g > 40000 && b > 40000 {
				c = color.Black
			}
			draw.Draw(img, image.Rect(x-5, y-5, x+5, y+5), &image.Uniform{c}, image.ZP, draw.Src)
		}
		var buff bytes.Buffer
		jpeg.Encode(&buff, img, &jpeg.Options{Quality: 100 - compression})
		lastScreen = "data:image/jpeg;base64," + base64.StdEncoding.EncodeToString(buff.Bytes())
		checkReduceFPS(imageStart)
		time.Sleep(frameTime)
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

func screen(w http.ResponseWriter, r *http.Request) {
	if !authenticate(w, r) {
		return
	}
	if checkSocketActive(w, r, true) {
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
			fmt.Println("readScreen:", err)
			c.Close()
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
					robotgo.Click("left", false)
				case "R":
					robotgo.Click("right", false)
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
		case "C":
			c.Close()
		case "O":
			switch string(string(message)[2]) {
				case "E":
					checkMobileMode(strings.Split(r.URL.Path, "_")[1], true)
				case "D":
					checkMobileMode(strings.Split(r.URL.Path, "_")[1], false)
			}
		}
	}
}

func checkMobileMode(id string, activate bool) {
	fmt.Println("mobile mode toggled in socket: " + id)
	intId, err := strconv.Atoi(id)
	if err == nil {
		mobileSockets[intId] = activate
	}
	for i := range mobileSockets {
		if mobileSockets[i] {
			mobileMode = true
			fmt.Println("mobile mode remains active")
			return
		}
	}
	mobileMode = false
	fmt.Println("mobile mode turned off")
}

func socketActivity(c *websocket.Conn, mt int, id string) {
	for {
		time.Sleep(frameTime)
		if len(lastScreen) > 0 {
			err := c.WriteMessage(mt, []byte(lastScreen))
			if err != nil {
				determineGlobalActivity()
				c.Close()
				break
			}
		}
		if !active {
			c.Close()
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
	if server {
		socket := strconv.Itoa(determineSocket())
		sockets := map[string]interface{}{
			"screen": "ws://" + r.Host + "/screen_" + socket,
		}
		scriptTemplate.Execute(w, sockets)
	} else {
		w.WriteHeader(503)
	}
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
	if speed {
		frameTime = 0
		return
	}
	if time.Now().UnixMilli()-start.UnixMilli() > frameTime.Milliseconds() {
		fps = fps / 2
	} else {
		fps = fps * 2
	}
	calculateFrameTime()
}

func determineSocket() int {
	for i := range sockets {
		if !sockets[i] {
			fmt.Println("assign socket: " + strconv.Itoa(i))
			return i
		}
	}
	sockets = append(sockets, false)
	mobileSockets = append(mobileSockets, false)
	fmt.Println("create and assign socket: " + strconv.Itoa(len(sockets)-1))
	m.HandleFunc("/screen_"+strconv.Itoa(len(sockets)-1), screen)
	return len(sockets) - 1
}

func activateSocket(id string) {
	fmt.Println("activate socket: " + id)
	intId, err := strconv.Atoi(id)
	if err == nil {
		sockets[intId] = true
	}
}

func deactivateSocket(id string) {
	fmt.Println("deactivate socket: " + id)
	intId, err := strconv.Atoi(id)
	if err == nil {
		sockets[intId] = false
	}
}

func checkSocketActive(w http.ResponseWriter, r *http.Request, writeHeader bool) bool {
	id, err := strconv.Atoi(strings.Split(r.URL.Path, "_")[1])
	if err == nil && len(sockets) > id {
		if sockets[id] {
			fmt.Println("socket already active: " + strconv.Itoa(id))
			if writeHeader {
				w.WriteHeader(503)
			}
			return true
		}
	}
	return false
}

func checkAvailableSockets() bool {
	activeSockets := 0
	for i := range sockets {
		if sockets[i] {
			activeSockets++
		}
		i++
	}
	return activeSockets == sessions
}

func determineGlobalActivity() {
	for i := range sockets {
		if sockets[i] && !active {
			active = true
			go makeImage()
			fmt.Println("activate render")
			return
		}
		if sockets[i] {
			return
		}
	}
	if active {
		fmt.Println("deactivate render")
		active = false
		if !speed {
			fmt.Println("cleanup endpoints")
			s.Shutdown(context.Background())
			server = false
		}
	}
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
var port int
var sessions int
var frameTime time.Duration
var lastScreen string
var compression int
var active = false
var m *http.ServeMux
var s http.Server
var sockets []bool
var mobileSockets []bool
var mobileMode = false
var speed bool
var server = false
