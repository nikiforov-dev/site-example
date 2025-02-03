.PHONY: start stop

build:
	CGO_ENABLED=1 go build -o myapp .

start:
	echo "Starting server..."; \
	nohup ./myapp > server.log 2>&1 &

