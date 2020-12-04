# goMirror

Small app written in go, which streams the screen of the current system on 
a web page in port 80.

Mouse usage is also possible, click on the edges of the screen to move the mouse
and on the center to perform a left click!

If you click on the capture keyboard checkbox, you can send single keystrokes to the 
remote machine!

Just run, access the system's IP
address through a browser and click start!

Libraries:

            WebSocket       - https://github.com/gorilla/websocket
            ScreenShot      - https://github.com/vova616/screenshot
            RobotGo         - https://github.com/go-vgo/robotgo
            Resize          - https://github.com/nfnt/resize
