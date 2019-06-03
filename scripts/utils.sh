#!/usr/bin/env bash

# shell utilities

function build() {
    VERSION_DATA=`cat ${PROJECT}/version.go`
    VERSION_DATA=${VERSION_DATA#*"Version = \""}
    VERSION=${VERSION_DATA%%[!0-9.]*}
    TARGET=${PROJECT}/dist/teaweb-tunnel-v${VERSION}
    EXT=""
    if [ ${GOOS} = "windows" ]
    then
        EXT=".exe"
    fi

    echo "[================ building ${GOOS}/${GOARCH}/v${VERSION}] ================]"

    echo "[goversion]using" `go version`
    echo "[create target directory]"

    if [ ! -d ${PROJECT}/dist ]
    then
		mkdir ${PROJECT}/dist
    fi

    if [ -d ${TARGET} ]
    then
        rm -rf ${TARGET}
    fi

    mkdir ${TARGET}
    mkdir ${TARGET}/bin
    mkdir ${TARGET}/configs
    mkdir ${TARGET}/logs

    echo "[build static file]"

    # build main & plugin
    go build -ldflags="-s -w" -o ${TARGET}/bin/teaweb-tunnel${EXT} ${PROJECT}/main/main.go

    echo "[copy files]"
    cp -R ${PROJECT}/main/configs/config.sample.yml ${TARGET}/configs/config.yml

    echo "[zip files]"
    cd ${TARGET}/../
    if [ -f teaweb-tunnel-${GOOS}-${GOARCH}-v${VERSION}.zip ]
    then
        rm -f teaweb-tunnel-${GOOS}-${GOARCH}-v${VERSION}.zip
    fi
    zip -r -X -q teaweb-tunnel-${GOOS}-${GOARCH}-v${VERSION}.zip  teaweb-tunnel-v${VERSION}/
    cd -

    echo "[clean files]"
    rm -rf ${TARGET}

    echo "[done]"
}