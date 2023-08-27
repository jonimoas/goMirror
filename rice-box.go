// Code generated by rice embed-go; DO NOT EDIT.
package main

import (
	"time"

	"github.com/GeertJohan/go.rice/embedded"
)

func init() {

	// define files
	file2 := &embedded.EmbeddedFile{
		Filename:    "index.tmpl",
		FileModTime: time.Unix(1681497946, 0),

		Content: string("\n<html>\n  <head>\n    <link rel=\"stylesheet\" href=\"style\">\n    <meta charset=\"utf-8\" />\n    <title>goMirror</title>\n  </head>\n  <body bgcolor=\"#000000\">\n    <button id=\"start\">Start</button>\n    <button id=\"end\">Stop</button>\n    <input\n      type=\"checkbox\"\n      id=\"capturekey\"\n      name=\"capturekey\"\n      value=\"capturekey\"\n    />\n    <label for=\"capturekey\">Capture Keyboard</label>\n    <input type=\"checkbox\" id=\"queuekey\" name=\"queuekey\" value=\"queuekey\" />\n    <label for=\"queuekey\">Queue Keyboard</label>\n    <button id=\"send\">Send Queue</button>\n    <br />\n    <img src=\"\" id=\"screen\" />\n  </body>\n  <script src=\"script\"></script>\n</html>\n\n"),
	}
	file3 := &embedded.EmbeddedFile{
		Filename:    "script.js",
		FileModTime: time.Unix(1693157159, 0),

		Content: string("var screenSocket;\nvar inputSocket;\ndocument.getElementById(\"start\").onclick = function (evt) {\n  var password = prompt(\"Enter Password\");\n  screenSocket = new WebSocket(\"{{.screen}}\" + \"?password=\" + password);\n  inputSocket = new WebSocket(\"{{.input}}\" + \"?password=\" + password);\n  screenSocket.onmessage = function (evt) {\n    document.getElementById(\"screen\").src = evt.data;\n    return false;\n  };\n  screenSocket.onopen = function (evt) {\n    screenSocket.send(\"go\");\n    return false;\n  };\n  return false;\n};\ndocument.getElementById(\"end\").onclick = function (evt) {\n  screenSocket.send(\"stop\");\n  screenSocket.close();\n  return false;\n};\ndocument.getElementById(\"send\").onclick = function (evt) {\n  inputSocket.send(\"K-W\");\n  document.getElementById(\"queuekey\").checked = false;\n  return false;\n};\ndocument.getElementById(\"screen\").onclick = function (evt) {\n  var move = false;\n  bounds = this.getBoundingClientRect();\n  var left = bounds.left;\n  var top = bounds.top;\n  var x = event.pageX - left;\n  var y = event.pageY - top;\n  var cw = this.clientWidth;\n  var ch = this.clientHeight;\n  var iw = this.naturalWidth;\n  var ih = this.naturalHeight;\n  var px = (x / cw) * iw;\n  var py = (y / ch) * ih;\n  if (px <= iw * 0.3) {\n    move = true;\n    inputSocket.send(\"M-M-L\");\n    if (px <= iw * 0.3) {\n      inputSocket.send(\"M-M-L\");\n    }\n  }\n  if (py <= ih * 0.3) {\n    move = true;\n    inputSocket.send(\"M-M-U\");\n    if (py <= ih * 0.15) {\n      inputSocket.send(\"M-M-U\");\n    }\n  }\n  if (px >= iw - 0.3 * iw) {\n    move = true;\n    inputSocket.send(\"M-M-R\");\n    if (px >= iw - 0.3 * iw) {\n      inputSocket.send(\"M-M-R\");\n    }\n  }\n  if (py >= ih - 0.3 * ih) {\n    move = true;\n    inputSocket.send(\"M-M-D\");\n    if (py >= ih - 0.3 * ih) {\n      inputSocket.send(\"M-M-D\");\n    }\n  }\n  if (!move) {\n    inputSocket.send(\"M-C-L\");\n  }\n};\ndocument.getElementById(\"screen\").oncontextmenu = function (evt) {\n  inputSocket.send(\"M-C-R\");\n};\ndocument.getElementById(\"queuekey\").onchange = function (evt) {\n  inputSocket.send(\"K-E\");\n};\nwindow.addEventListener(\n  \"keydown\",\n  function (event) {\n    if (event.defaultPrevented) {\n      return;\n    }\n    if (document.getElementById(\"capturekey\").checked) {\n      if (document.getElementById(\"queuekey\").checked) {\n        inputSocket.send(\"K-Q-\" + event.key);\n        event.preventDefault();\n        return false;\n      } else {\n        inputSocket.send(\"K-T-\" + event.key);\n        event.preventDefault();\n        return false;\n      }\n    }\n  },\n  true\n);\nwindow.addEventListener(\"beforeunload\", function(e){\n  screenSocket.send(\"stop\");\n  screenSocket.close();\n});\n"),
	}
	file4 := &embedded.EmbeddedFile{
		Filename:    "style.css",
		FileModTime: time.Unix(1681497946, 0),

		Content: string("label {\n    color: white;\n}\n\n#screen {\n    touch-action: manipulation;\n    transition-timing-function: linear;\n}"),
	}

	// define dirs
	dir1 := &embedded.EmbeddedDir{
		Filename:   "",
		DirModTime: time.Unix(1693159680, 0),
		ChildFiles: []*embedded.EmbeddedFile{
			file2, // "index.tmpl"
			file3, // "script.js"
			file4, // "style.css"

		},
	}

	// link ChildDirs
	dir1.ChildDirs = []*embedded.EmbeddedDir{}

	// register embeddedBox
	embedded.RegisterEmbeddedBox(`frontend`, &embedded.EmbeddedBox{
		Name: `frontend`,
		Time: time.Unix(1693159680, 0),
		Dirs: map[string]*embedded.EmbeddedDir{
			"": dir1,
		},
		Files: map[string]*embedded.EmbeddedFile{
			"index.tmpl": file2,
			"script.js":  file3,
			"style.css":  file4,
		},
	})
}
