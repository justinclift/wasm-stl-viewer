'use strict';

const WASM_URL = 'wasm.wasm';
var wasm;

// Canvas resize callback
function canvasResize() {
    wasm.exports.canvasResize();
}

function sliderChangeX(evt) {
    wasm.exports.sliderChangeX(evt.currentTarget.value);
}

function sliderChangeY(evt) {
    wasm.exports.sliderChangeY(evt.currentTarget.value);
}

function sliderChangeZ(evt) {
    wasm.exports.sliderChangeZ(evt.currentTarget.value);
}

function zoomChange(evt) {
    wasm.exports.zoomChange(evt.deltaY);
}

// Render one frame of the animation
function renderFrame(evt) {
    wasm.exports.renderFrame(evt);
}

// Callback handlers for model uploading
// function uploadAbort(something) {
//     console.log("uploadAbort: JS value = " + something);
//     wasm.exports.uploadAborted(something);
// }
//
// function uploadError(something) {
//     console.log("uploadError: JS value = " + something);
//     wasm.exports.uploadError(something);
// }
//
function inputEvent(something) {
    console.log(something);
// function uploaded(something, something2) {
    console.log("inputEvent: JS value = " + something);
    wasm.exports.uploaded(something);
}

// function uploaded(something) {
// // function uploaded(something, something2) {
//     console.log("uploaded: JS value = " + something);
//     wasm.exports.uploaded(something);
// }

function uploading(something) {
    console.log("uploading: JS value = " + something);
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
      document.getElementById("upload").addEventListener("input", inputEvent);
      document.getElementById("upload").addEventListener("change", uploading);

      const reader = new FileReader();
      // reader.onload = uploaded(img);
      // reader.onload = (function(aImg) { return console.log(); })();
      // reader.readAsDataURL();
      reader.onload = (function(aImg) { return function(e) { console.log("Testing..."); }; })();
      // reader.onload = (function(aImg) { return function(e) { aImg.src = e.target.result; }; })(img);
      // reader.addEventListener("load", uploaded);
      // reader.addEventListener("error", uploadError);
      // reader.addEventListener("abort", uploadAbort);
      // document.get("FileReader").addEventListener("load", uploaded);
        // newReader := js.Global().Get("FileReader")
        //reader.Call("addEventListener", "load", uploadedCallback)
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
