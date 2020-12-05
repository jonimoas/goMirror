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
	"net"
	"net/http"
	"os"
	"strings"
	ttemplate "text/template"

	"github.com/go-vgo/robotgo"
	"github.com/gorilla/websocket"
	"github.com/vova616/screenshot"
)

var keyBuffer []string

func main() {
	flag.Parse()
	log.SetFlags(0)
	http.HandleFunc("/screen", screen)
	http.HandleFunc("/input", input)
	http.HandleFunc("/script", script)
	http.HandleFunc("/", home)
	fmt.Println("IP Addresses:")
	host, _ := os.Hostname()
	addrs, _ := net.LookupIP(host)
	for _, addr := range addrs {
		if ipv4 := addr.To4(); ipv4 != nil {
			fmt.Println("IPv4: ", ipv4)
		}
	}
	log.Fatal(http.ListenAndServe(":80", nil))
}

func makeImage() string {
	img, err := screenshot.CaptureScreen()
	x, y := robotgo.GetMousePos()
	c := color.White
	pointer := image.Rect(x-5, y-5, x+5, y+5)
	draw.Draw(img, pointer, &image.Uniform{c}, image.ZP, draw.Src)
	if err != nil {
		panic(err)
	}
	var buff bytes.Buffer
	jpeg.Encode(&buff, img, &jpeg.Options{Quality: 50})
	encodedString := base64.StdEncoding.EncodeToString(buff.Bytes())
	htmlImage := "data:image/jpeg;base64," + encodedString
	return htmlImage
}

var upgrader = websocket.Upgrader{}

func input(w http.ResponseWriter, r *http.Request) {
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Print("upgrade:", err)
		return
	}
	for {
		mt, message, err := c.ReadMessage()
		if err != nil {
			log.Println("read:", err)
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
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Print("upgrade:", err)
		return
	}
	for {
		mt, message, err := c.ReadMessage()
		if err != nil {
			log.Println("read:", err)
			break
		}
		if string(message) == "go" {
			for {
				err = c.WriteMessage(mt, []byte(makeImage()))
				if err != nil {
					log.Println("write:", err)
					break
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

var homeTemplate, errH = htemplate.ParseGlob("*.tmpl")
var scriptTemplate, errS = ttemplate.ParseGlob("*.js")
