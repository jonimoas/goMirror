package main

import (
	"bytes"
	"encoding/base64"
	"flag"
	"html/template"
	"image/png"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/kbinani/screenshot"
)

func main() {
	flag.Parse()
	log.SetFlags(0)
	http.HandleFunc("/echo", echo)
	http.HandleFunc("/", home)
	log.Fatal(http.ListenAndServe(":80", nil))

}

func makeImage() string {
	bounds := screenshot.GetDisplayBounds(0)
	img, err := screenshot.CaptureRect(bounds)
	if err != nil {
		panic(err)
	}
	var buff bytes.Buffer
	png.Encode(&buff, img)
	encodedString := base64.StdEncoding.EncodeToString(buff.Bytes())
	htmlImage := "data:image/png;base64," + encodedString
	return htmlImage
}

var upgrader = websocket.Upgrader{}

func echo(w http.ResponseWriter, r *http.Request) {
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
		for {
			err = c.WriteMessage(mt, []byte(makeImage()))
			if err != nil {
				log.Println("write:", err)
				break
			}
		}
	}

}

func home(w http.ResponseWriter, r *http.Request) {
	homeTemplate.Execute(w, "ws://"+r.Host+"/echo")
}

var homeTemplate, err = template.ParseGlob("*.tmpl")
