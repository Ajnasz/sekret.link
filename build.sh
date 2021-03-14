#!/bin/sh

set -eu

build() {
	VERSION=$(git describe --tags)
	BUILD=$(date +%FT%T%z)
	ARCHLIST="$1"
	OSLIST="$2"
	STORAGE="$3"
	BUILD_DIR="$4"

	echo "$VERSION"
	echo "$BUILD"

  mkdir -p $BUILD_DIR

	for os in $OSLIST
	do
		for arch in $ARCHLIST
		do
			for storage in $STORAGELIST
			do
				echo "building $os.$arch.$storage"
				GOOS="$os" GOARCH="$arch" go build -tags "$storage" -ldflags "-w -s -X main.version=${VERSION} -X main.build=${BUILD}" -o "$BUILD_DIR/sekret.link.$os.$arch.$storage"
			done
		done
	done
}

remove() {
  BUILD_DIR="$1"
	for os in $OSLIST
	do
		for arch in $ARCHLIST
		do
			for storage in $STORAGELIST
			do
				if [ -f "$BUILD_DIR/sekret.link.$os.$arch.$storage" ]
				then
					echo "removing $os.$arch.$storage"
					rm "$BUILD_DIR/sekret.link.$os.$arch.$storage"
				fi
			done
		done
	done
}

REMOVE=0
OSLIST="linux"
ARCHLIST="amd64 386"
STORAGELIST="postgres"
BUILD_DIR="./build"
STORAGE=""
BUILD=0

if [ $# -lt 1 ];then
  echo "Command required, available commands are \"test\" and \"build\"" >&2
  exit 1
fi

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
		echo "Invalid command, available commands are \"test\" and \"build\"" >&2
		exit 1
		;;
esac

while getopts "ra:o:s:b:" opt
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
    "b")
      BUILD_DIR="$OPTARG"
      ;;
		[?])
			exit 1
			;;
	esac
done

if [ $REMOVE -eq 1 ];then
	remove "$BUILD_DIR"
else
	build "$ARCHLIST" "$OSLIST" "$STORAGE" "$BUILD_DIR"
fi
