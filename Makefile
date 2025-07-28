default: build

build:
	go build -ldflags="-w -s" -trimpath -o build/manweb

clean:
	rm -rf build

install_bin: ./build/manweb
	sudo apt-get install -y mandoc
	sudo sh ./extras/preinst.sh
	sudo install ./build/manweb /usr/bin/manweb
	sudo install ./extras/manweb-passwd /usr/bin/manweb-passwd
	sudo install ./extras/manweb.service /etc/manweb/manweb.service
	sudo install ./extras/manweb.conf /etc/manweb/manweb.conf
	sudo sh ./extras/postinstall.sh

