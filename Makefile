build:
	go build -o glazel ./cmd/glazel

install:
	go build -o glazeld ./cmd/glazeld
	go build -o glazel-agent ./cmd/glazel-agent
	go build -o glazel ./cmd/glazel

	sudo mv glazel /usr/local/bin/
	sudo mv glazel-agent /usr/local/bin/
	sudo mv glazeld /usr/local/bin/
