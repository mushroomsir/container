build: 
	GOOS=linux GOARCH=amd64 go build main.go
run: 
	sudo ./main run bash
config:
	mkdir -p rootfs
	sudo tar -xvf rootfs.tar -C rootfs