default: build
build:
	go build -ldflags="-w -s" -trimpath -o build/manhttpd

build_deb: ./build/manhttpd
	cd extras && nfpm pkg --packager deb --config nfpm.yaml --target ../build/manhttpd.deb

install_deb: ./build/manhttpd.deb
	sudo apt-get install -y ./build/manhttpd.deb

install_bin: ./build/manhttpd
	sudo apt-get install -y mandoc
	sudo sh ./extras/preinst.sh
	sudo install ./build/manhttpd /usr/bin/manhttpd
	sudo install ./extras/manhttpd-passwd /usr/bin/manhttpd-passwd
	sudo install ./extras/manhttpd.service /etc/manhttpd/manhttpd.service
	sudo install ./extras/manhttpd.conf /etc/manhttpd/manhttpd.conf
	sudo sh ./extras/postinstall.sh

