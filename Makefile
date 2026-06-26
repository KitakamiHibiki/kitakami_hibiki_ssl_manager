.PHONY: build run dev clean all

all: build-frontend build

build:
	go build -o bin/server ./cmd/server

build-frontend:
	cd web && npm run build

run: build-frontend
	./bin/server

dev:
	go run ./cmd/server

frontend:
	cd web && npm run dev

clean:
	rm -rf bin/ web/dist/
