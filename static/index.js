// How long points are displayed (in milliseconds)
const DISPLAY_TIME = 1000 * 30;
var circles = [];

// distroID, color, counter, isPresent
const distros = [
  ["alpine", "#cd5700", 0, 0],
  ["archlinux", "#f0d4f0", 0, 0],
  ["archlinux32", "#d8bfd8", 0, 0],
  ["artix-linux", "#de6fa1", 0, 0],
  ["blender", "#eb7700", 0, 0],
  ["centos", "#0abab5", 0, 0],
  ["clonezilla", "#e08d3c", 0, 0],
  ["cpan", "#dbd7d2", 0, 0],
  ["cran", "#eee600", 0, 0],
  ["ctan", "#ff6347", 0, 0],
  ["cygwin", "#746cc0", 0, 0],
  ["debian", "#ffc87c", 0, 0],
  ["debian-cd", "#ffc87c", 0, 0],
  ["eclipse", "#00755e", 0, 0],
  ["freebsd", "#deaa88", 0, 0],
  ["gentoo", "#30d5c8", 0, 0],
  ["gentoo-portage", "#30d5c8", 0, 0],
  ["gparted", "#a0d6b4", 0, 0],
  ["ipfire", "#7c4848", 0, 0],
  ["isabelle", "#8a496b", 0, 0],
  ["linux", "#66023c", 0, 0],
  ["linuxmint", "#62b858", 0, 0],
  ["manjaro", "#d9004c", 0, 0],
  ["msys2", "#8878c3", 0, 0],
  ["odroid", "#536895", 0, 0],
  ["openbsd", "#ffb300", 0, 0],
  ["opensuse", "#3cd070", 0, 0],
  ["parrot", "#ff6fff", 0, 0],
  ["raspbian", "#120a8f", 0, 0],
  ["RebornOS", "#4166f5", 0, 0],
  ["ros", "#635147", 0, 0],
  ["sabayon", "#ffddca", 0, 0],
  ["serenity", "#5b92e5", 0, 0],
  ["slackware", "#b78727", 0, 0],
  ["slitaz", "#ff6", 0, 0],
  ["tdf", "#00a500", 0, 0],
  ["templeos", "#efcc00", 0, 0],
  ["ubuntu", "#ffd300", 0, 0],
  ["ubuntu-cdimage", "#ffdd32", 0, 0],
  ["ubuntu-ports", "#ccaa00", 0, 0],
  ["ubuntu-releases", "#ccb028", 0, 0],
  ["videolan", "#0014a8", 0, 0],
  ["voidlinux", "#00", 0, 0],
  ["zorinos", "#fc6c85", 0, 0],
];

function WebSocketTest() {
  let xhr = new XMLHttpRequest();
  xhr.open("GET", "/register");
  xhr.setRequestHeader("Access-Control-Allow-Headers", "*");
  xhr.setRequestHeader("Access-Control-Allow-Origin", "*");
  xhr.send();
  xhr.onload = function () {
    console.log("id:", xhr.response);
    var id = xhr.response;

    if ("WebSocket" in window) {
      // Let us open a web socket
      var url = "ws://" + location.host + "/socket/" + id;
      var ws = new WebSocket(url);

      ws.onmessage = function (evt) {
        var reader = new FileReader();
        var msg;

        reader.readAsArrayBuffer(evt.data);
        reader.addEventListener("loadend", function(e)
        {
          buffer = new Uint8Array(reader.result);
          // console.log(String.fromCharCode(convertToDecimal(buffer[2])));

          let distro = parseFloat(buffer[0], 2);

          let longByte = buffer.slice(1, 9);
          let latByte = buffer.slice(9, 17);

          let latBuf = new ArrayBuffer(8);
          let latView = new DataView(latBuf);
          latByte.forEach(function (b, i) {
            latView.setUint8(i, b);
          });

          let lat = latView.getFloat64(0, true);

          let longBuf = new ArrayBuffer(8);
          let longView = new DataView(longBuf);
          longByte.forEach(function (b, i) {
            longView.setUint8(i, b);
          });

          let long = longView.getFloat64(0, true);

          let x = (lat + 180) / 360;
          let y = (90 - long) / 180;
          distros[distro][2] += 1;
          // Add new data points to the front of the list
          circles.unshift([x, y, distro, new Date().getTime()]);

        });
      };

      ws.onclose = function () {
        // websocket is closed.
        alert("Connection is closed\nrefresh to connect");
      };
    } else {
      // The browser doesn't support WebSocket
      alert(
        "This site will not work since\nyour browser does not support websockets"
      );
    }
  };
}

window.onload = async function () {
  WebSocketTest();

  const canvas = document.getElementById("myCanvas");
  const ctx = canvas.getContext("2d");
  const img = document.getElementById("map");

  window.onresize = function () {
    canvas.width = window.innerWidth;
    canvas.height = window.innerHeight;
  };

  window.onresize();

  while (true) {
    let checkTime = new Date().getTime();

    ctx.globalAlpha = 1;
    ctx.drawImage(img, 0, 0, canvas.width, canvas.height);

    for (let i = 0; i < circles.length; i++) {
      circle = circles[i];
      distros[circle[2]][3] = 1;

      // Time difference
      const delta = checkTime - circles[i][3];

      // Remove old data points
      if (delta > DISPLAY_TIME) {
        console.log("REMOVING ", circles.length - i);
        // We know all future indexes are older
        circles = circles.slice(0, i);
        break;
      }

      ctx.fillStyle = distros[circle[2]][1];
      ctx.beginPath();
      ctx.globalAlpha = 1 - delta / DISPLAY_TIME;
      ctx.arc(
        circle[0] * canvas.width,
        circle[1] * canvas.height,
        2.0,
        0,
        2 * Math.PI,
        false
      );
      ctx.closePath();
      ctx.fill();
    }

    ctx.beginPath();
    let incX = 0;
    let incY = 0;
    let startX = 10;
    let startY = canvas.height * 0.44;
    let maxPerColumn = (canvas.height * (0.9 - 0.44)) / 15;
    let numberOfEntries = distros.map((d) => d[3]).reduce((a, b) => a + b);

    if (numberOfEntries == 0) {
      await new Promise((r) => setTimeout(r, 15));
      continue;
    }

    let numberOfRows = Math.ceil(numberOfEntries / maxPerColumn);

    // Show rectangle
    let height = Math.min(canvas.height * (0.9 - 0.44), 15 * numberOfEntries);
    let width = numberOfRows * 130;
    ctx.globalAlpha = 1;
    ctx.fillStyle = "#282828";
    ctx.rect(5, startY - 40, width, height + 45);
    ctx.fill();

    // "Legend"
    ctx.fillStyle = "white";
    ctx.textAlign = "center";
    ctx.fillText("Legend", width * 0.5, startY - 20);

    // Print each visible distro
    ctx.font = "15px Arial";
    ctx.textAlign = "left";
    const sorted = [...distros].sort((a, b) => b[2] - a[2]);
    for (let i = 0; i < sorted.length; i++) {
      if (sorted[i][3] == 1) {
        if (startY + incY > canvas.height * 0.9) {
          incY = 0;
          incX += 130;
        }
        ctx.fillStyle = sorted[i][1];
        ctx.fillText(sorted[i][0], startX + incX, startY + incY);
        incY += 15;
        sorted[i][3] = 0;
      }
    }

    // Run around 60 fps
    await new Promise((r) => setTimeout(r, 15));
  }
};
