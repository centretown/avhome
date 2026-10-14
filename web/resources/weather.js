function toggleClass(id, toggle, klass) {
  toggle = !toggle;
  swapClass(id, klass);
  return toggle;
}

const cwLoc = new Map();
function drawCurrentItems(locID) {
  cwLoc.get(locID).drawItems();
}
function toggleCurrentItem(locID, itemID) {
  cwLoc.get(locID).toggleItem(itemID);
}
function toggleHistory(locID, id) {
  cwLoc.get(locID).toggleDetail(id);
}

const hrLoc = new Map();
function drawHourlyItems(locID) {
  hrLoc.get(locID).drawItems();
}
function toggleHourlyItem(locID, itemID) {
  hrLoc.get(locID).toggleItem(itemID);
}
function toggleHourlyDetail(locID, id) {
  console.log(locId, id);
  hrLoc.get(locID).toggleDetail(id);
}
// document.getElementById("hourly-date").innerHTML = showFullDate(
//     new Date(),
// );

function locationItem(prefix, canvasID, codes, items) {
  let obj = {
    hide: true,
    canvasID: canvasID,
    prefix: prefix,
    codes: codes,
    items: items,
    drawItems: function () {
      clearCanvas(this.canvasID);
      showMinMax(this.canvasID);
      for (let [_, item] of this.items) {
        if (item.selected) {
          item.draw(this.canvasID, this.codes);
        }
      }
    },
    toggleItem: function (itemID) {
      console.log(itemID);
      let item = this.items.get(itemID);
      item.selected = toggleClass(itemID, item.selected, "title-selected");
      swapClass(itemID + "-max", "hide");
      swapClass(itemID + "-min", "hide");
      this.drawItems();
    },
    toggleDetail: function (id) {
      this.hide = !this.hide;
      swapClass(id, "hide");
    },
  };
  return obj;
}

const rain = "rgba(31, 144, 255, 255)";
const snow = "rgba(255, 255, 255, 255)";

function codeToColor(code) {
  if ((code >= 71 && code <= 77) || (code >= 85 && code <= 86)) {
    return snow;
  }
  return rain;
}

function roundedRect(ctx, x, y, width, height, radius) {
  ctx.beginPath();
  ctx.moveTo(x, y + radius);
  ctx.arcTo(x, y + height, x + radius, y + height, radius);
  ctx.arcTo(x + width, y + height, x + width, y + height - radius, radius);
  ctx.arcTo(x + width, y, x + width - radius, y, radius);
  ctx.arcTo(x, y, x, y + radius, radius);
  // ctx.fillStyle = "rgba(0,0,255, 1)";
  // ctx.fill();
  ctx.stroke();
}

const x_origin = 2;
const y_origin = 4;

function showMinMaxRect(canvasId) {
  let canvas = document.getElementById(canvasId);
  let ctx = canvas.getContext("2d");

  ctx.setLineDash([]);
  ctx.lineWidth = 0.1;
  ctx.strokeStyle = "white";

  let yStep = (canvas.height - 2 * y_origin) / 10;
  let xStep = (canvas.width - 2 * x_origin) / 24;
  for (let y = y_origin; y < canvas.height - yStep - 1; y += yStep) {
    for (let x = x_origin; x < canvas.width - xStep - 1; x += xStep) {
      ctx.beginPath();
      ctx.strokeRect(x, y, x + xStep - 1, y + yStep - 1);
      ctx.stroke();
    }
  }
}

function showMinMax(canvasId) {
  let canvas = document.getElementById(canvasId);
  let ctx = canvas.getContext("2d");

  ctx.beginPath();
  // ctx.setLineDash([]);
  ctx.lineWidth = 0.1;
  ctx.strokeStyle = "white";

  let yneg = false;
  let xneg = false;

  let half = 0.5;
  let yStep = (canvas.height - 2 * y_origin) / 10;
  for (let y = y_origin + yStep; y < canvas.height - yStep; y += yStep) {
    if (y % 2 != 0) {
      y += (yneg) ? -half : half;
      yneg = !yneg;
    }
    ctx.moveTo(x_origin, y);
    ctx.lineTo(canvas.width - x_origin, y);
  }

  let xStep = canvas.width / 24;

  for (let x = xStep + x_origin; x < canvas.width; x += xStep) {
    if (x % 2 != 0) {
      x += (xneg) ? -half : half;
      xneg = !xneg;
    }
    ctx.moveTo(x, y_origin);
    ctx.lineTo(x, canvas.height - y_origin);
  }

  ctx.stroke();
}

function clearCanvas(canvasId) {
  let canvas = document.getElementById(canvasId);
  let ctx = canvas.getContext("2d");
  ctx.beginPath();
  ctx.fillStyle = "rgba(0, 0, 0, 1)";
  ctx.clearRect(0, 0, canvas.width, canvas.height);
}

function showBars(canvasId, values, codes) {
  let canvas = document.getElementById(canvasId);
  let ctx = canvas.getContext("2d");
  let xStep = canvas.width / values.length;
  let yStep = 5;
  let precision = (7 * 0.1) / values.length;
  for (let i = 0; i < values.length; i++) {
    let x = i * xStep + x_origin;
    let yStart = canvas.height - y_origin;
    let y = yStart;
    ctx.beginPath();
    ctx.strokeStyle = codeToColor(codes[i]);
    ctx.lineWidth = 1;
    ctx.setLineDash([]);

    for (let value = values[i]; value > 0.0; value = value - precision) {
      ctx.moveTo(x, y);
      ctx.lineTo(x + xStep - 2, y);
      x = i * xStep + x_origin;
      y -= yStep;
      if (y <= y_origin) {
        yStart--;
        y = yStart;
      }
    }

    ctx.stroke();
  }
}

function showGraph(canvasId, color, min, max, values, lineWidth = 1) {
  let canvas = document.getElementById(canvasId);
  let ctx = canvas.getContext("2d");
  let xStep = (canvas.width - 2 * x_origin) / (values.length - 1);
  let height = canvas.height - 2 * y_origin;
  let yStep = height / values.length;
  if (max > min) yStep = height / (max - min);

  ctx.beginPath();
  ctx.strokeStyle = color;
  ctx.lineWidth = lineWidth;
  ctx.setLineDash([]);
  let x = x_origin;
  let y = height;
  // ctx.font = "10px sans-serif";

  for (let val of values) {
    y = (max - val) * yStep + y_origin;
    if (x == 0) ctx.moveTo(x, y);
    else ctx.lineTo(x, y);
    x += xStep;
  }
  ctx.stroke();
}
