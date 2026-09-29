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
  let element = document.getElementById(id);
  if (element === undefined) {
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

const w3_hide = "w3-hide";
const icon_down = "arrow_drop_down";
const icon_right = "arrow_right";
const element_not_found = "element not found";
const invalid_arguments = "invalid arguments";

function toggle_fold(id, icon_id) {
  const elt = document.getElementById(id);
  const icon_elt = document.getElementById(icon_id);
  if (elt && icon_elt) {
    if (elt.className.indexOf(w3_hide) == -1) {
      elt.classList.add(w3_hide);
      icon_elt.innerHTML = icon_right;
    } else {
      elt.classList.remove(w3_hide);
      icon_elt.innerHTML = icon_down;
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
