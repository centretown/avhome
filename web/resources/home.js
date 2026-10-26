function swapClass(id, klass1) {
  console.log(id, klass1);
  const element = document.getElementById(id);
  if (element.className.includes(klass1)) {
    element.classList.remove(klass1);
  } else {
    element.classList.add(klass1);
  }
  return;
}

function swapContent(id) {
  let element = document.getElementById(id);
  if (element === undefined || element == null) {
    console.log(id, "undefined");
    return;
  }

  if (element.parentElement.id === "side-list") {
    const swap = contentView.firstElementChild;
    listView.append(swap);
    contentView.append(element);
  } else {
    const swap = listView.firstElementChild;
    contentView.append(swap);
    listView.append(element);
  }
}

const clockFmt = new Intl.DateTimeFormat("en-US", {
  // weekday: "short",
  // month: "short",
  // day: "numeric", // dateStyle: "full",
  hour: "numeric",
  minute: "numeric",
  timeZone: "America/New_York",
});

const clockList = [];
addClockList("clock", clockFmt);
// id of element, timeFormat Intl.DateTimeFormat
function addClockList(id, timeFormat) {
  clockList.push({ id, fmt: timeFormat });
}

function startClock() {
  const today = new Date();
  for (item of clockList) {
    document.getElementById(item.id).innerHTML = item.fmt.format(today);
  }
  setTimeout(startClock, 1000 * (60 - today.getSeconds()));
}

const w3_hide = "hide";
const icon_down = "arrow_drop_down";
const icon_right = "arrow_right";
const element_not_found = "element not found";
const invalid_arguments = "invalid arguments";

const toggle_map = new Map();

function addToggleItem(id, toggle_item) {
  toggle_map.set(id, toggle_item);
}

function toggle_fold(id, icon_id) {
  const elt = document.getElementById(id);
  const icon_elt = document.getElementById(icon_id);
  const toggle_item = toggle_map.get(id);
  if (!toggle_item) {
    console.log("toggle_item undefined", id);
  }

  if (elt && icon_elt) {
    if (elt.className.indexOf(w3_hide) == -1) {
      elt.classList.add(w3_hide);
      icon_elt.innerHTML = icon_right;
      if (toggle_item) {
        toggle_item.close();
      }
    } else {
      elt.classList.remove(w3_hide);
      icon_elt.innerHTML = icon_down;
      if (toggle_item) {
        toggle_item.open();
      }
    }
  }
}

function get_folded(id) {
  const elt = document.getElementById(id);
  if (elt && elt.className.indexOf(w3_hide) == -1) {
    return icon_right;
  }
  return icon_down;
}

function showFullDate(dt) {
  let options = {
    timeStyle: "short",
    dateStyle: "full",
    timeZone: "America/New_York",
  };
  let fmt = new Intl.DateTimeFormat("en-US", options);
  return fmt.format(dt);
}

// thanks to pjdietz github gist
function pad(n) {
  return n < 10 ? "0" + n : n;
}

function timezoneOffset(offset) {
  if (offset === 0) {
    return "Z";
  }
  let sign = offset > 0 ? "-" : "+";
  offset = Math.abs(offset);
  return sign + pad(Math.floor(offset / 60)) + ":" + pad(offset % 60);
}

function local_date_time_format(dt) {
  return (
    dt.getFullYear() +
    "-" +
    pad(dt.getMonth() + 1) +
    "-" +
    pad(dt.getDate()) +
    "T" +
    pad(dt.getHours()) +
    ":" +
    pad(dt.getMinutes())
  );
}

function rfc3339(dt) {
  return (
    dt.getFullYear() +
    "-" +
    pad(dt.getMonth() + 1) +
    "-" +
    pad(dt.getDate()) +
    "T" +
    pad(dt.getHours()) +
    ":" +
    pad(dt.getMinutes()) +
    ":" +
    pad(dt.getSeconds()) +
    timezoneOffset(dt.getTimezoneOffset())
  );
}
