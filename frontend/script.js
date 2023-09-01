var screenSocket;
var screenSocket;
document.getElementById("start").onclick = function (evt) {
  var password = prompt("Enter Password");
  try{
    screenSocket = new WebSocket("{{.screen}}" + "?password=" + password);
    screenSocket.onmessage = function (evt) {
      document.getElementById("screen").src = evt.data;
      return false;
    };
    screenSocket.onopen = function (evt) {
      screenSocket.send("go");
      return false;
    };
    screenSocket.onerror = function (evt) {
      location.reload();
    };
    return false;
  } catch(e) {
    location.reload();
  }
};
document.getElementById("end").onclick = function (evt) {
  screenSocket.send("stop");
  screenSocket.close();
  return false;
};
document.getElementById("send").onclick = function (evt) {
  screenSocket.send("K-W");
  document.getElementById("queuekey").checked = false;
  return false;
};
document.getElementById("screen").onclick = function (evt) {
  if (document.getElementById("mobilemode").checked) {
    var move = false;
    bounds = this.getBoundingClientRect();
    var left = bounds.left;
    var top = bounds.top;
    var x = event.pageX - left;
    var y = event.pageY - top;
    var cw = this.clientWidth;
    var ch = this.clientHeight;
    var iw = this.naturalWidth;
    var ih = this.naturalHeight;
    var px = (x / cw) * iw;
    var py = (y / ch) * ih;
    if (px <= iw * 0.3) {
      move = true;
      screenSocket.send("M-M-L");
      if (px <= iw * 0.15) {
        screenSocket.send("M-M-L");
      }
    }
    if (py <= ih * 0.3) {
      move = true;
      screenSocket.send("M-M-U");
      if (py <= ih * 0.15) {
        screenSocket.send("M-M-U");
      }
    }
    if (px >= iw - 0.3 * iw) {
      move = true;
      screenSocket.send("M-M-R");
      if (px >= iw - 0.15 * iw) {
        screenSocket.send("M-M-R");
      }
    }
    if (py >= ih - 0.3 * ih) {
      move = true;
      screenSocket.send("M-M-D");
      if (py >= ih - 0.15 * ih) {
        screenSocket.send("M-M-D");
      }
    }
    if (!move) {
      screenSocket.send("M-C-L");
    }
  } else {
    x = evt.offsetX;
    y = evt.offsetY;
    screenSocket.send("M-A-" + x + "-" + y);
    screenSocket.send("M-C-L");
  }
};
document.getElementById("screen").oncontextmenu = function (evt) {
  screenSocket.send("M-C-R");
};
document.getElementById("queuekey").onchange = function (evt) {
  screenSocket.send("K-E");
};
document.getElementById("mobilemode").onchange = function (evt) {
  if (document.getElementById("mobilemode").checked) {
    document.getElementById("screencontainer").style.position = "fixed";
    document.getElementById("screencontainer").style.height = "100%";
    document.getElementById("screencontainer").style.width = "100%";
    document.getElementById("screen").style.height = "100%";
    document.getElementById("screen").style.width = "100%";
    document.getElementById("screen").style["object-fit"] = "scale-down";
    screenSocket.send("O-E");
  } else {
    document.getElementById("screencontainer").style.position = "";
    document.getElementById("screencontainer").style.height = "";
    document.getElementById("screencontainer").style.width = "";
    document.getElementById("screen").style.height = "";
    document.getElementById("screen").style.width = "";
    document.getElementById("screen").style["object-fit"] = "";
    screenSocket.send("O-D");
  }

};
window.addEventListener(
  "keydown",
  function (event) {
    if (event.defaultPrevented) {
      return;
    }
    if (document.getElementById("capturekey").checked) {
      if (document.getElementById("queuekey").checked) {
        screenSocket.send("K-Q-" + event.key);
        event.preventDefault();
        return false;
      } else {
        screenSocket.send("K-T-" + event.key);
        event.preventDefault();
        return false;
      }
    }
  },
  true
);
window.addEventListener("beforeunload", function(e){
  screenSocket.send("stop");
  screenSocket.close();
});
