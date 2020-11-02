#!/bin/sh

build() {
	VERSION=$(git describe --tags)
	BUILD=$(date +%FT%T%z)
	ARCHLIST="$1"
	OSLIST="$2"
	STORAGE="$3"
	echo $VERSION
	echo $BUILD
	for os in $OSLIST
	do
		for arch in $ARCHLIST
		do
			for storage in $STORAGELIST
			do
				echo "building $os.$arch.$storage"
				GOOS=$os GOARCH=$arch go build -tags $storage -ldflags "-w -s -X main.version=${VERSION} -X main.build=${BUILD}" -o "sekret.link.$os.$arch.$storage"
			done
		done
	done
}

remove() {
	for os in $OSLIST
	do
		for arch in $ARCHLIST
		do
			for storage in $STORAGELIST
			do
				if [ -f "sekret.link.$os.$arch.$storage" ]
				then
					echo "removing $os.$arch.$storage"
					rm "sekret.link.$os.$arch.$storage"
				fi
			done
		done
	done
}

REMOVE=0
OSLIST="linux darwin freebsd"
ARCHLIST="amd64 386"
STORAGELIST="postgres redis sqlite"
BUILD=0

subcommand="$1"
shift
case "$subcommand" in
	"test")
		go test -v -tags test
		return
		;;
	"build")
		BUILD=1;
		;;
	*)
		echo "Invalid command, available commands are \"test\" and \"build\""
		exit 1
		;;
esac

while getopts "tra:o:s:" opt
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
			STORAGELIST="$OPTARG"
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
