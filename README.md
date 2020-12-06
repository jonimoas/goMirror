# goMirror

Small app written in go, which streams the screen of the current system on 
a web page in port 80.

Mouse usage is also possible, click on the edges of the screen to move the mouse
and on the center to perform a left click, or a right click! (works with long press on mobile)

If you click on the capture keyboard checkbox, you can send single keystrokes to the 
remote machine!

If you enable the queue, all keys will be stored and will either be pressed simulteneously
when you press the button, or cleared if you uncheck the box!

Just run, access the system's IP address through a browser, click start and enter the password
that is displayed on the console window of the host computer! 

Libraries:

            WebSocket       - https://github.com/gorilla/websocket
            ScreenShot      - https://github.com/vova616/screenshot
            RobotGo         - https://github.com/go-vgo/robotgo

![Screenshot](/screenshot.png)