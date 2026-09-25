function addReader(video_id, pathName, message_id) {
  const video_element = document.getElementById(video_id);
  const message_element = document.getElementById(message_id);
  return new MediaMTXWebRTCReader({
    url: "http://192.168.10.7:8889/" + pathName + "/whep",
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

function swapClass(id, klass1) {
  const element = document.getElementById(id);
  if (element.className.includes(klass1)) {
    element.classList.remove(klass1);
  } else {
    element.classList.add(klass1);
  }
  return;
}


function swapContentVideo(id) {
  const element = document.getElementById(id);
  const previous = contentView.firstElementChild;
  contentView.clientWidth, contentView.clientHeight
  contentView.append(element);
  listView.append(previous);
}

function startClock() {
  const today = new Date();
  let clockFmt = new Intl.DateTimeFormat("en-US", {
    weekday: "short",
    month: "short",
    day: "numeric", // dateStyle: "full",
    hour: "numeric",
    minute: "numeric",
    timeZone: "America/New_York",
  });

  document.getElementById("clock").innerHTML = clockFmt.format(today);
  setTimeout(startClock, 1000 * (60 - today.getSeconds()));
}
