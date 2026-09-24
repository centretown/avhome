

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
      video_element.controls = true
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
