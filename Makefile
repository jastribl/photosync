export MAKEFLAGS="-j 8"

build_all: cacheitems createalbum drive2photos findallmissinglocal findallmissingphotos labelphotos spacesaver

cacheitems:
	go build -o bin/$@ cmd/$@/main.go

createalbum:
	go build -o bin/$@ cmd/$@/main.go

drive2photos:
	go build -o bin/$@ cmd/$@/main.go

findallmissinglocal:
	go build -o bin/$@ cmd/$@/main.go

findallmissingphotos:
	go build -o bin/$@ cmd/$@/main.go

labelphotos:
	go build -o bin/$@ cmd/$@/main.go

spacesaver:
	go build -o bin/$@ cmd/$@/main.go