build_all:
	go build cmd/cacheitems/main.go;
	go build cmd/createalbum/main.go;
	go build cmd/drive2photos/main.go;
	go build cmd/findallmissinglocal/main.go;
	go build cmd/findallmissingphotos/main.go;
	go build cmd/labelphotos/main.go;
	go build cmd/spacesaver/main.go;
	rm main;
