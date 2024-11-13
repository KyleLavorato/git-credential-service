#!/bin/bash

usage() {
    echo ''
    echo 'Usage: build.sh <component>'
    echo ''
    echo 'Options:'
    echo '  -u | Do unit testing'
    echo '  -v | Verbose output'
    echo '  -l | Skip linting'
    echo '  -a | Build for ARM (Override auto-detect)'
    echo '  -i | Build for x86 (Override auto-detect)'
    echo ''
    echo 'Components:'
    echo '  all | Build the complete package'
    echo '  test | Run all unit testing and generate junit output'
    echo '  clean | Delete all build components'
    exit 1
}

if [ -z "$1" ]; then
    usage
fi

ZIPNAME=git-credential-service-lambda.zip

set -eu

dotest=0
verbose=0
skipLint=0
intelBuild=0
armBuild=0
while getopts ":vuail" opt; do
    case $opt in
        (u) dotest=1;;
        (v) verbose=1;;
        (l) skipLint=1;;
        (a) armBuild=1;;
        (i) intelBuild=1;;
        (*) usage
    esac
done
shift "$((OPTIND - 1))"

if [[ $verbose == 1 ]]; then
    set -eux
fi

GOARCH=
if [ $(uname -m) == 'aarch64' ] || [ $(uname -m) == 'arm64' ] || [ $armBuild == 1 ]; then
    echo "Building for ARM"
    intelBuild=0
    armBuild=1
    GOARCH=arm64
else
    echo "Building for x86"
    intelBuild=1
    armBuild=0
    GOARCH=amd64
fi

cwd=$(pwd)
out=$cwd/out
publish=$cwd/publish

mkdir -p $out
rm -rf $publish || true
mkdir -p $publish/bin

buildGoLambda() {
    go env -w GOFLAGS='-buildvcs=false'

    # To use the AWS provided.al2 Go runtime, the binary must be named `bootstrap`
    GO_OUTPUT_BIN=bootstrap

    go fmt ./...
    GOOS=linux \
        GOARCH=$GOARCH \
        go build -tags lambda.norpc -ldflags="-s -w" -o $out/bootstrap main.go
    go vet ./...
    go tool buildid $out/bootstrap

    if [[ $skipLint == 0 ]]; then
        # do this after builds since it could fail for build reasons but without the explanations
        if type "golangci-lint" &>/dev/null ; then
            golangci-lint run --exclude-dirs="(go/pkg/|go/src/)"
        fi
    fi

    pushd $out
        zip --symlinks $publish/bin/$ZIPNAME \
            bootstrap
    popd
    ls -la $publish/bin

    if [[ $dotest == 1 ]]; then
        doGoTestLocal
    fi
}

doGoTestLocal() {
    echo "Starting Go Testing"
    rm -rf $cwd/test-result || true
    mkdir -p $cwd/test-result
    go test -timeout 1m ./... 2>&1> test-result/go-test-result || true # Don't exit on error
    if grep -q "FAIL" test-result/go-test-result; then
        echo "Go Testing FAILED"
        echo "See test-result/go-test-result for more details"
        exit 1
    else
        echo "Go Testing PASSED"
    fi
}

doGoTestJenkins() {
    pushd $cwd/test-result
        GOLANG_LOG_TO_CONSOLE=true \
            go test -timeout 5m -v -cover -coverprofile cp.out ../... 2>&1 > report || true # Do not exit on test failure
        cat report | go-junit-report > report.xml
        gocover-cobertura < cp.out > coverage.xml
        go tool cover -func cp.out > coverage.txt
        go tool cover -html=cp.out -o coverage.html
    popd
}

doClean() {
    rm -rf $out
    rm -rf $publish
}

echo "Build Selection: $1"
case $1 in
    all)
        buildGoLambda
        ;;
    clean)
        doClean
        ;;
    test)
        rm -rf $cwd/test-result || true
        mkdir -p $cwd/test-result
        doGoTestJenkins
        ;;
    *)
        echo "$1 is not a valid build option"
        exit 1
        ;;
esac
