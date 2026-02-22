build:
	go build -o glazel ./cmd/glazel

install:
	go build -o glazel ./cmd/glazel
	sudo mv glazel /usr/local/bin/
