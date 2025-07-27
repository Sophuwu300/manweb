default: build

build:
	go build -ldflags="-w -s" -trimpath -o build/manweb

clean:
	rm -rf build

build_deb:
	cd extra && nfpm pkg --packager deb --config nfpm.yaml --target ../build/manweb.deb

install_deb: ./build/manweb.deb
	sudo apt-get install -y ./build/manweb.deb

install_bin: ./build/manweb
	sudo apt-get install -y mandoc
	sudo sh ./extras/preinst.sh
	sudo install ./build/manweb /usr/bin/manweb
	sudo install ./extras/manweb-passwd /usr/bin/manweb-passwd
	sudo install ./extras/manweb.service /etc/manweb/manweb.service
	sudo install ./extras/manweb.conf /etc/manweb/manweb.conf
	sudo sh ./extras/postinstall.sh

