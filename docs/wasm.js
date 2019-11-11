'use strict';

const WASM_URL = 'wasm.wasm';
var wasm;

// Canvas resize callback
function canvasResize() {
    wasm.exports.canvasResize();
}

function sliderChangeX(evt) {
    let val = evt.currentTarget.value;
    console.log("JS value = " + val);
    wasm.exports.sliderChangeX(val);
}

function sliderChangeY(evt) {
    wasm.exports.sliderChangeY(evt.currentTarget.value);
}

function sliderChangeZ(evt) {
    wasm.exports.sliderChangeZ(evt.currentTarget.value);
}

function zoomChange(evt) {
    console.log(evt);
    wasm.exports.zoomChange(evt);
}

// Render one frame of the animation
function renderFrame(evt) {
    wasm.exports.renderFrame(evt);
}

function uploading(something) {
    wasm.exports.uploading(something);
}

// Load and run the wasm
function init() {
  const go = new Go();
  if ('instantiateStreaming' in WebAssembly) {
    WebAssembly.instantiateStreaming(fetch(WASM_URL), go.importObject).then(function (obj) {
      wasm = obj.instance;
      go.run(wasm);

      // Set up wasm event handlers
      document.addEventListener("resize", canvasResize);
      document.addEventListener("wheel", zoomChange);
      document.getElementById("speedSliderX").addEventListener("input", sliderChangeX);
      document.getElementById("speedSliderY").addEventListener("input", sliderChangeY);
      document.getElementById("speedSliderZ").addEventListener("input", sliderChangeZ);
      document.getElementById("upload").addEventListener("change", uploading);

    })
  } else {
    fetch(WASM_URL).then(resp =>
      resp.arrayBuffer()
    ).then(bytes =>
      WebAssembly.instantiate(bytes, go.importObject).then(function (obj) {
        wasm = obj.instance;
        go.run(wasm);

        // Set up wasm event handlers
        document.addEventListener("resize", canvasResize);
        document.addEventListener("wheel", zoomChange);
        document.getElementById("speedSliderX").addEventListener("input", sliderChangeX);
        document.getElementById("speedSliderY").addEventListener("input", sliderChangeY);
        document.getElementById("speedSliderZ").addEventListener("input", sliderChangeZ);
        document.getElementById("upload").addEventListener("change", uploading);

      })
    )
  }
}

init();
