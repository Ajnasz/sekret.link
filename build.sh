#!/bin/sh

build() {
	VERSION=$(git describe --tags)
	BUILD=$(date +%FT%T%z)
	ARCHLIST="$1"
	OSLIST="$2"
	STORAGE="$3"
	echo $VERSION
	echo $BUILD
	echo $STORAGE
	for os in $OSLIST
	do
		for arch in $ARCHLIST
		do
			echo "building $os.$arch.$STORAGE"
			GOOS=$os GOARCH=$arch go build -tags $STORAGE -ldflags "-w -s -X main.version=${VERSION} -X main.build=${BUILD}" -o "sekret.link.$os.$arch.$STORAGE"
		done
	done
}

remove() {
	for os in $OSLIST
	do
		for arch in $ARCHLIST
		do
			if [ -f "sekret.link.$os.$arch.$STORAGE" ]
			then
				echo "removing $os.$arch.$STORAGE"
				rm "sekret.link.$os.$arch.$STORAGE"
			fi
		done
	done
}

REMOVE=0
OSLIST="linux darwin freebsd"
ARCHLIST="amd64 386"
STORAGE="postgres"

while getopts "ra:o:" opt
do
	case "$opt" in
		"r")
			REMOVE=1
			;;
		"a")
			ARCHLIST="$OPTARG"
			;;
		"o")
			OSLIST="$OPTARG"
			;;
		"s")
			STORAGE="$OPTARG"
			;;
		[?])
			exit 1
			;;
	esac
done

if [ $REMOVE -eq 1 ];then
	remove
else
	build "$ARCHLIST" "$OSLIST" "$STORAGE"
fi
