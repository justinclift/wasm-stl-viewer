package main

import (
	"encoding/base64"
	"errors"
	"math"
	"strconv"
	"strings"
	"syscall/js"

	"github.com/justinclift/wasm-stl-viewer/models"
	"github.com/justinclift/wasm-stl-viewer/renderer"
)

var (
	gl      js.Value
	//glTypes gltypes.GLTypes
)

//// BUFFERS + SHADERS ////
// Shamelessly copied from https://www.tutorialspoint.com/webgl/webgl_cube_rotation.htm //
var verticesNative = []float32{
	-1, -1, -1, 1, -1, -1, 1, 1, -1, -1, 1, -1,
	-1, -1, 1, 1, -1, 1, 1, 1, 1, -1, 1, 1,
	-1, -1, -1, -1, 1, -1, -1, 1, 1, -1, -1, 1,
	1, -1, -1, 1, 1, -1, 1, 1, 1, 1, -1, 1,
	-1, -1, -1, -1, -1, 1, 1, -1, 1, 1, -1, -1,
	-1, 1, -1, -1, 1, 1, 1, 1, 1, 1, 1, -1,
}
var colorsNative = []float32{
	5, 3, 7, 5, 3, 7, 5, 3, 7, 5, 3, 7,
	1, 1, 3, 1, 1, 3, 1, 1, 3, 1, 1, 3,
	0, 0, 1, 0, 0, 1, 0, 0, 1, 0, 0, 1,
	1, 0, 0, 1, 0, 0, 1, 0, 0, 1, 0, 0,
	1, 1, 0, 1, 1, 0, 1, 1, 0, 1, 1, 0,
	0, 1, 0, 0, 1, 0, 0, 1, 0, 0, 1, 0,
}
var indicesNative = []uint32{
	0, 1, 2, 0, 2, 3, 4, 5, 6, 4, 6, 7,
	8, 9, 10, 8, 10, 11, 12, 13, 14, 12, 14, 15,
	16, 17, 18, 16, 18, 19, 20, 21, 22, 20, 22, 23,
}

const vertShaderCode = `
attribute vec3 position;
uniform mat4 Pmatrix;
uniform mat4 Vmatrix;
uniform mat4 Mmatrix;
attribute vec3 color;
varying vec3 vColor;

void main(void) {
	gl_Position = Pmatrix*Vmatrix*Mmatrix*vec4(position, 1.);
	vColor = color;
}
`
const fragShaderCode = `
precision mediump float;
varying vec3 vColor;
void main(void) {
	gl_FragColor = vec4(vColor, 1.);
}
`

var reader js.Value
var render renderer.Renderer
var speedSliderXValue js.Value
var speedSliderYValue js.Value
var speedSliderZValue js.Value
var canvasElement js.Value
var currentZoom float32 = 3

func uploading(this js.Value, args []js.Value) interface{} {
	files := this.Get("files")
	file := files.Index(0)
	currentFileName := file.Get("name").String()
	println("Uploading", currentFileName)
	reader.Call("readAsDataURL", file)
	return nil
}

func parseBase64File(input string) (output []byte, err error) {
	searchString := "base64,"
	index := strings.Index(input, searchString)
	if index < 0 {
		err = errors.New("Error opening file")
		return
	}
	sBuffer := input[index+len(searchString):]
	return base64.StdEncoding.DecodeString(sBuffer)
}

func uploaded(this js.Value, args []js.Value) interface{} {
	println("Finished uploading")
	result := args[0].Get("target").Get("result").String()
	func() {
		defer func() {
			if r := recover(); r != nil {
				println("Recovered in upload", r)
				js.Global().Call("alert", "Failed to parse file")
			}
		}()
		uploadedFile, err := parseBase64File(result)
		if err != nil {
			panic(err)
		}
		stlSolid, err := models.NewSTL(uploadedFile)
		if err != nil {
			js.Global().Call("alert", "Could not parse file")
		}
		vert, colors, indices := stlSolid.GetModel()
		modelSize := getMaxScalar(vert)
		currentZoom = modelSize * 3
		render.SetZoom(currentZoom)
		render.SetModel(colors, vert, indices)
	}()
	return nil
}

func getMaxScalar(vertices []float32) float32 {
	var max float32
	for baseIndex := 0; baseIndex < len(vertices); baseIndex += 3 {
		testScale := scalar(vertices[baseIndex], vertices[baseIndex], vertices[baseIndex])
		if testScale > max {
			max = testScale
		}
	}
	return max
}

func scalar(x float32, y float32, z float32) float32 {
	xy := math.Sqrt(float64(x*x + y*y))
	return float32(math.Sqrt(xy*xy + float64(z*z)))
}

func uploadError(this js.Value, args []js.Value) interface{} {
	println("Uploading Error")
	return nil
}

func uploadAborted(this js.Value, args []js.Value) interface{} {
	println("Upload Aborted")
	return nil
}

func main() {
	println("Returned normally from f.")

	// Init Canvas stuff
	doc := js.Global().Get("document")

	canvasResizeCallback := js.FuncOf(canvasResize)
	canvasElement = doc.Call("getElementById", "gocanvas")
	js.Global().Get("window").Call("addEventListener", "resize", canvasResizeCallback)

	width := canvasElement.Get("clientWidth").Int()
	height := canvasElement.Get("clientHeight").Int()
	canvasElement.Set("width", width)
	canvasElement.Set("height", height)
	upload := doc.Call("getElementById", "upload")
	newReader := js.Global().Get("FileReader")
	reader = newReader.New()

	sliderSpeedXCallback := js.FuncOf(sliderChangeX)
	speedSliderX := doc.Call("getElementById", "speedSliderX")
	speedSliderX.Call("addEventListener", "input", sliderSpeedXCallback)
	speedSliderXValue = doc.Call("getElementById", "speedSliderXValue")

	sliderSpeedYCallback := js.FuncOf(sliderChangeY)
	speedSliderY := doc.Call("getElementById", "speedSliderY")
	speedSliderY.Call("addEventListener", "input", sliderSpeedYCallback)
	speedSliderYValue = doc.Call("getElementById", "speedSliderYValue")

	sliderSpeedZCallback := js.FuncOf(sliderChangeZ)
	speedSliderZ := doc.Call("getElementById", "speedSliderZ")
	speedSliderZ.Call("addEventListener", "input", sliderSpeedZCallback)
	speedSliderZValue = doc.Call("getElementById", "speedSliderZValue")

	zoomChangeCallback := js.FuncOf(zoomChange)
	js.Global().Get("window").Call("addEventListener", "wheel", zoomChangeCallback)

	uploadCallback := js.FuncOf(uploading)
	uploadedCallback := js.FuncOf(uploaded)
	errorUploadCallback := js.FuncOf(uploadError)
	uploadAbortedCallback := js.FuncOf(uploadAborted)
	defer uploadCallback.Release()
	defer uploadedCallback.Release()
	defer errorUploadCallback.Release()
	defer uploadAbortedCallback.Release()
	reader.Call("addEventListener", "load", uploadedCallback)
	reader.Call("addEventListener", "error", errorUploadCallback)
	reader.Call("addEventListener", "abort", uploadAbortedCallback)
	upload.Call("addEventListener", "change", uploadCallback)

	gl = canvasElement.Call("getContext", "webgl")
	if gl == js.Undefined() {
		gl = canvasElement.Call("getContext", "experimental-webgl")
	}
	if gl == js.Undefined() {
		js.Global().Call("alert", "browser might not support webgl")
		return
	}

	config := renderer.InitialConfig{
		Width:              width,
		Height:             height,
		SpeedX:             0.5,
		SpeedY:             0.3,
		SpeedZ:             0.2,
		Colors:             colorsNative,
		Vertices:           verticesNative,
		Indices:            indicesNative,
		FragmentShaderCode: fragShaderCode,
		VertexShaderCode:   vertShaderCode,
	}
	var err error
	render, err = renderer.NewRenderer(gl, config)
	if err != nil {
		js.Global().Call("alert", "Cannot load webgl "+err.Error())
		return
	}
	render.SetZoom(currentZoom)
	defer render.Release()

	x, y, z := render.GetSpeed()
	speedSliderX.Set("value", strconv.FormatFloat(float64(x), 'f', 0, 32))
	speedSliderXValue.Set("innerHTML", strconv.FormatFloat(float64(x), 'f', 0, 32))
	speedSliderY.Set("value", strconv.FormatFloat(float64(y), 'f', 0, 32))
	speedSliderYValue.Set("innerHTML", strconv.FormatFloat(float64(y), 'f', 0, 32))
	speedSliderZ.Set("value", strconv.FormatFloat(float64(z), 'f', 0, 32))
	speedSliderZValue.Set("innerHTML", strconv.FormatFloat(float64(z), 'f', 0, 32))

	var renderFrame js.Func
	renderFrame = js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		render.Render(this, args)
		js.Global().Call("requestAnimationFrame", renderFrame)
		return nil
	})
	js.Global().Call("requestAnimationFrame", renderFrame)
}

func canvasResize(this js.Value, args []js.Value) interface{} {
	width := canvasElement.Get("clientWidth").Int()
	height := canvasElement.Get("clientHeight").Int()
	canvasElement.Set("width", width)
	canvasElement.Set("height", height)
	render.SetSize(height, width)
	return nil
}

func sliderChangeX(this js.Value, args []js.Value) interface{} {
	var speed float32
	sSpeed := this.Get("value").String()
	//fmt.Sscan(sSpeed, &speed) // TODO: Figure out a replacement for fmt.Sscan()
	render.SetSpeedX(speed)
	speedSliderXValue.Set("innerHTML", sSpeed)
	return nil
}

func sliderChangeY(this js.Value, args []js.Value) interface{} {
	var speed float32
	sSpeed := this.Get("value").String()
	//fmt.Sscan(sSpeed, &speed)
	render.SetSpeedY(speed)
	speedSliderYValue.Set("innerHTML", sSpeed)
	return nil
}

func sliderChangeZ(this js.Value, args []js.Value) interface{} {
	var speed float32
	sSpeed := this.Get("value").String()
	//fmt.Sscan(sSpeed, &speed)
	render.SetSpeedZ(speed)
	speedSliderZValue.Set("innerHTML", sSpeed)
	return nil
}

func zoomChange(this js.Value, args []js.Value) interface{} {
	deltaY := args[0].Get("deltaY").Float()
	deltaScale := 1 - (float32(deltaY) * 0.001)
	currentZoom *= deltaScale
	render.SetZoom(currentZoom)
	return nil
}
