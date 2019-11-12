all: wasm-stl-viewer

wasm-stl-viewer: docs/wasm.wasm server.go
	go build -o wasm-stl-viewer server.go

docs/wasm.wasm: main.go color/color.go color/gradient.go color/interpolation.go models/model.go models/stl.go renderer/renderer.go
	tinygo build -target wasm -no-debug -o docs/wasm.wasm main.go
	wasm2wat docs/wasm.wasm -o docs/wasm.wat
	wat2wasm docs/wasm.wat -o docs/wasm.wasm
	rm -f docs/wasm.wat

run: docs/wasm.wasm wasm-stl-viewer
	./wasm-stl-viewer

clean:
	rm -f server.wasm docs/wasm.wasm
