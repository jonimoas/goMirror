# goMirror

Small app written in go, which streams the screen of the current system on 
a web page.

Mouse usage possible, either by clicking on the edges of the screen to move the mouse
and on the center to perform a left click, or a right click! (works with long press on mobile, if movile mode is on),
or by simply clicking on the desired point of the screen (if mobile mode is off)

If you click on the capture keyboard checkbox, you can send single keystrokes to the 
remote machine!

If you enable the queue, all keys will be stored and will either be pressed simulteneously
when you press the button, or cleared if you uncheck the box!

Just run, access the system's IP address through a browser, click start and enter the password
that is displayed on the console window of the host computer!

FPS is calculated in real time to avoid overloading the CPU

Usage of ./goMirror:

  -fps int

        the framerate at which the app will start (default 60)

  -pass string

        the desired password, will generate one by default (default "Generated")

  -port int

        the port that te app will be hosted on (default 80)

  -sessions int

        the maximum number of websocket endpoints to be created (default 5)



NOTE: if you want to rebuild the frontend, you need to use rice and the rice embed-go command.

Libraries:

            WebSocket       - https://github.com/gorilla/websocket
            ScreenShot      - https://github.com/kbinani/screenshot
            RobotGo         - https://github.com/go-vgo/robotgo
            Rice            - https://github.com/GeertJohan/go.rice

![Screenshot](/screenshot.png)
