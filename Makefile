.PHONY: start stop

start:
	echo "Starting server..."; \
	nohup go run . > server.log 2>&1 &

