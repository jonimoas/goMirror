package main

import (
	"bytes"
	"encoding/base64"
	"flag"
	"html/template"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"log"
	"net/http"

	"github.com/go-vgo/robotgo"
	"github.com/gorilla/websocket"
	"github.com/nfnt/resize"
	"github.com/vova616/screenshot"
)

func main() {
	flag.Parse()
	log.SetFlags(0)
	http.HandleFunc("/screen", screen)
	http.HandleFunc("/input", input)
	http.HandleFunc("/", home)
	log.Fatal(http.ListenAndServe(":80", nil))
	robotgo.MoveMouseSmooth(100, 200, 1.0, 100.0)
}

func makeImage() string {
	img, err := screenshot.CaptureScreen()
	width := uint((float64(img.Bounds().Dx()) * 0.5))
	height := uint((float64(img.Bounds().Dy()) * 0.5))
	x, y := robotgo.GetMousePos()
	c := color.White
	pointer := image.Rect(x-5, y-5, x+5, y+5)
	draw.Draw(img, pointer, &image.Uniform{c}, image.ZP, draw.Src)
	img.Set(x, y, c)
	resizedImg := resize.Resize(width, height, image.Image(img), resize.Lanczos3)
	if err != nil {
		panic(err)
	}
	var buff bytes.Buffer
	png.Encode(&buff, resizedImg)
	encodedString := base64.StdEncoding.EncodeToString(buff.Bytes())
	htmlImage := "data:image/png;base64," + encodedString
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
		log.Printf("recv: %s", message)
		if string(string(message)[0]) == "M" {
			switch string(string(message)[2]) {
			case "U":
				robotgo.MoveSmoothRelative(0, -10, 3.0, 30.0)
			case "D":
				robotgo.MoveSmoothRelative(0, 10, 3.0, 30.0)
			case "L":
				robotgo.MoveSmoothRelative(-10, 0, 3.0, 30.0)
			case "R":
				robotgo.MoveSmoothRelative(10, 0, 3.0, 30.0)
			case "C":
				robotgo.MouseClick("left", true)
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
		log.Printf("recv: %s", message)
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
	sockets := map[string]interface{}{
		"screen": "ws://" + r.Host + "/screen",
		"input":  "ws://" + r.Host + "/input",
	}
	homeTemplate.Execute(w, sockets)
}

var homeTemplate, err = template.ParseGlob("*.tmpl")
