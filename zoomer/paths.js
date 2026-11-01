const mtx_paths = [];
const mtx_readers = [];
const mtx_players = new Map();

mtx_paths.push({
  host: "10.0.0.7",
  name: "cam1",
  start: "2026-10-17T14:48:46+00:00",
  end: "2026-10-17T14:48:46+00:00",
  duration: 60,
});

function addPath(path) {
  mtx_readers.push(
    addReader(
      path.host,
      path.name + "-video",
      path.name,
      path.name + "-message",
    ),
  );
  const toggleItem = {
    name: path.name,
    id: path.name + "-player",
    open: function () {
      let videoPlayer = document.getElementById(this.id);
      if (videoPlayer) {
        videoPlayer.play();
      }
      console.log("open-" + this.name);
    },
    close: function () {
      let videoPlayer = document.getElementById(this.id);
      if (videoPlayer) {
        videoPlayer.pause();
      }
      console.log("close-" + this.name);
    },
  };

  let start = new Date();
  let end = new Date();
  start.setSeconds(-toggleItem.duration);
  end.setSeconds(0);
  toggleItem.start = rfc3339(start);
  toggleItem.end = rfc3339(end);

  const toggleId = path.name + "-player-fold";
  addToggleItem(toggleId, toggleItem);
}

function addPlayer(path) {
  start = "";
  end = "";
  duration = "";
  mtx_players.push({
    host: path.host,
    name: path.name,
    id: "player-" + path.name,
    src: "player-source-" + path.name,
    url:
      "http://" +
      path.host +
      ":9996/get?path=" +
      path.name +
      "&start=" +
      start +
      "&duration=" +
      duration,
    query:
      "http://" +
      path.host +
      ":9996/list?path=" +
      path.name +
      "&start=" +
      start +
      "&end=" +
      end,
  });
}

function loadPaths() {
  for (let path of mtx_paths) {
    console.log(path.host, path.name);
    addPath(path);
  }
}

function addReader(host, video_id, pathName, message_id) {
  const video_element = document.getElementById(video_id);
  const message_element = document.getElementById(message_id);
  return new MediaMTXWebRTCReader({
    url: "http://" + host + ":8889/" + pathName + "/whep",
    user: "", // fill if needed
    pass: "", // fill if needed
    token: "", // fill if needed
    onError: (err) => {
      message_element.innerText = err;
    },
    onTrack: (evt) => {
      video_element.controls = true;
      video_element.srcObject = evt.streams[0];
      message_element.innerText = "";
    },
    onDataChannel: (evt) => {
      evt.channel.binaryType = "arraybuffer";
      evt.channel.onmessage = (evt) => {
        console.log("data channel message", evt.data);
      };
    },
  });
}

function addPlayerTime(name) {
  const cur = document.getElementById(name + "-player-time");
  let val = new Date();
  val.setSeconds(-60 * 15);
  // dateControl.value = local_date_time_format(val);
  let min = new Date();
  min.setHours(-7 * 24);
  let max = new Date();
  cur.value = local_date_time_format(val);
  cur.min = local_date_time_format(min);
  cur.max = local_date_time_format(max);
  cur.step = "900";
}

let processor = {
  timerCallback: function () {
    if (this.video.paused || this.video.ended) {
      return;
    }
    this.computeFrame();
    let self = this;
    setTimeout(function () {
      self.timerCallback();
    }, 0);
  },

  doLoad: function () {
    this.video = document.getElementById("cam1-video");
    this.zoom = document.getElementById("cam1-canvas");
    this.ctxZoom = this.zoom.getContext("2d");

    this.width = 0;
    this.height = 0;
    this.wwidth = 0;
    this.wheight = 0;
    this.zoomLevel = 0;
    this.mouseX = 0;
    this.mouseY = 0;
    this.scale = 1;

    let self = this;
    this.video.addEventListener(
      "play",
      function () {
        self.width = self.video.videoWidth;
        self.height = self.video.videoHeight;
        self.timerCallback();
        console.log("playing...");
      },
      false,
    );
    self.video.addEventListener("wheel", (event) => {
      event.preventDefault();
      this.scale += event.deltaY * -0.005;
      this.scale = Math.min(Math.max(1, this.scale), 8);
      console.log(this.scale);
    });
    self.video.addEventListener("mousemove", () => {
      self.mouseX = Math.floor(
        (event.layerX * self.width) / self.video.offsetWidth,
      );
      self.mouseY = Math.floor(
        (event.clientY * self.height) / self.video.offsetHeight,
      );
      console.log(
        self.mouseX,
        self.mouseY,
        self.video.offsetWidth,
        self.video.offsetHeight,
      );
    });
  },

  computeFrame: function () {
    const sx = this.mouseX;
    const sy = this.mouseY;
    // this.ctxZoom.drawImage(this.video, start, 0, this.width, this.height);
    this.ctxZoom.drawImage(
      this.video,
      sx,
      sy,
      this.width,
      this.height,
      0,
      0,
      this.width,
      this.height,
    );
    return;
  },
};

document.addEventListener("DOMContentLoaded", () => {
  processor.doLoad();
});
