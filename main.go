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
	flag.Parse()
	log.SetFlags(0)
	http.HandleFunc("/screen", screen)
	http.HandleFunc("/input", input)
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
	fmt.Println("Password: " + password)
	var err error
	if len(os.Args) > 1 {
		fps, err = strconv.Atoi(os.Args[1])
		if err != nil {
			fps = 60
		}
	} else {
		fps = 60
	}
	maxfps = fps
	calculateFrameTime()
	log.Fatal(http.ListenAndServe(":80", nil))
}

func makeImage() {
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
		if string(message) == "go" {
			for {
				go makeImage()
				time.Sleep(frameTime)
				if lastScreen != "" {
					err := c.WriteMessage(mt, []byte(lastScreen))
					if err != nil {
						fmt.Println("write:", err)
						break
					}
				}
			}
		}

	}
}

func home(w http.ResponseWriter, r *http.Request) {
	homeTemplate.Execute(w, nil)
}

func script(w http.ResponseWriter, r *http.Request) {
	sockets := map[string]interface{}{
		"screen": "ws://" + r.Host + "/screen",
		"input":  "ws://" + r.Host + "/input",
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
	calculateFrameTime()
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
var password = randSeq(6)
var fps int
var maxfps int
var frameTime time.Duration
var lastScreen = ""
