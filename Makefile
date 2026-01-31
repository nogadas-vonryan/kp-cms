.PHONY: build build-frontend build-panel

# Build both UIs (required for embed) then the Wails binary
build: build-frontend build-panel
	wails build

build-frontend:
	export VITE_API_URL= && cd frontend && npm run build

build-panel:
	cd panel && npm run build
