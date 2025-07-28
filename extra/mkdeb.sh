#!/bin/bash

dir="$PWD"

goarch() {
	[[ -n "GOARCH" ]] && go build -ldflags="-w -s" -trimpath -o "$dir/build/manweb_$GOARCH" "$dir/main.go" ;
}
gobuild() {
	case "$1" in
		"amd"|"amd64"|"x86"|"x86_64")
			export GOARCH="amd64" ;
			echo building "build/manweb_$GOARCH" ;
			goarch ;;
		"arm"|"arm64")
			export GOARCH="arm64" ;
			echo building "build/manweb_$GOARCH" ;
			goarch ;;
		*)
			echo "invalid arch: $i" ;
			echo "valid values: amd64, arm64" ;
			return 1;;
	esac
	return 0
}

setVersion() {
  if [[ -z "$1" ]]; then
    echo "invalid version: $1" ;
    return 1 ;
  fi;
  local uwu="$1" ;
  uwu=$(printf "%d.%d.%d\n" ${uwu//./ } 2> /dev/null)
  if [[ $? -ne 0 ]]; then
    echo "invalid version: $1" ;
    return 1 ;
  fi
  for i in {amd64,arm64}; do
    sed --in-place "s/^version:.*\$/version: \"$uwu\"/g" "$dir/extra/nfpm_$i.yaml" ;
  done
  return 0 ;
}

mkDeb() {
  cd "$dir/extra";
  vers=$(grep 'version:' nfpm_amd64.yaml)
  vers=${vers:10:-1}
  for i in {amd64,arm64}; do
    gobuild "$i" && \
    nfpm pkg --packager deb --config "nfpm_$i.yaml" --target "$dir/build/manweb_${vers}_${i}.deb" && \
    echo "manweb_${vers}_${i}.deb created" || \
    echo "failed to build deb for $i" ;
  done
}


setVersion "$1" && mkDeb
