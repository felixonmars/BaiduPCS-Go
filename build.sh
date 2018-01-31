name="BaiduPCS-Go"
version="v3.2.1"

output="out/"

Build() {
    echo "Building $1..."
    export GOOS=$2 GOARCH=$3 GO386=sse2 CGO_ENABLED=0
    if [ $2 = "windows" ];then
        go build -ldflags "-s -w" -o "$output/$1/$name.exe"
    else
        go build -ldflags "-s -w" -o "$output/$1/$name"
    fi

    Pack $1
}

ArmBuild() {
    echo "Building $1..."
    export GOOS=$2 GOARCH=$3 GOARM=$4 CGO_ENABLED=1
    go build -ldflags '-s -w -linkmode=external -extldflags=-pie' -o "$output/$1/$name"
    if [ $2 = "darwin" -a $3 = "arm64" ];then
        ldid -S "$output/$1/$name"
    fi

    Pack $1
}

# 打包
Pack() {
    # rice 打包
    rice -i github.com/iikira/BaiduPCS-Go/pcsweb append --exec "$output/$1/$name"

    mkdir "$output/$1/download"
    cp README.md "$output/$1"

    cd $output
    zip -q -r "$1.zip" "$1"
    cd ..
}

# android
export NDK_INSTALL=$ANDROID_NDK_ROOT/bin
# CC=$NDK_INSTALL/arm-linux-androideabi-4.9/bin/arm-linux-androideabi-gcc ArmBuild $name-$version"-android-16-armv5" android arm 5
# CC=$NDK_INSTALL/arm-linux-androideabi-4.9/bin/arm-linux-androideabi-gcc ArmBuild $name-$version"-android-16-armv6" android arm 6
CC=$NDK_INSTALL/arm-linux-androideabi-4.9/bin/arm-linux-androideabi-gcc ArmBuild $name-$version"-android-16-armv7" android arm 7
CC=$NDK_INSTALL/aarch64-linux-android-4.9/bin/aarch64-linux-android-gcc ArmBuild $name-$version"-android-21-arm64" android arm64 7
CC=$NDK_INSTALL/i686-linux-android-4.9/bin/i686-linux-android-gcc ArmBuild $name-$version"-android-16-386" android 386 7
CC=$NDK_INSTALL/x86_64-linux-android-4.9/bin/x86_64-linux-android-gcc ArmBuild $name-$version"-android-21-amd64" android amd64 7

# ios 
CC=/usr/local/go/misc/ios/clangwrap.sh ArmBuild $name-$version"-darwin-ios-5.0-armv7" darwin arm 7
CC=/usr/local/go/misc/ios/clangwrap.sh ArmBuild $name-$version"-darwin-ios-5.0-arm64" darwin arm64 7

# OS X / macOS
Build $name-$version"-darwin-osx-amd64" darwin amd64
# Build $name-$version"-darwin-osx-386" darwin 386

# Windows
Build $name-$version"-windows-x86" windows 386
Build $name-$version"-windows-x64" windows amd64

# Linux
Build $name-$version"-linux-386" linux 386
Build $name-$version"-linux-amd64" linux amd64
Build $name-$version"-linux-arm" linux arm
Build $name-$version"-linux-arm64" linux arm64
# Build $name-$version"-linux-mips" linux mips
# Build $name-$version"-linux-mips64" linux mips64
# Build $name-$version"-linux-mipsel" linux mipsle
# Build $name-$version"-linux-mips64el" linux mips64le
# Build $name-$version"-linux-ppc64" linux ppc64
# Build $name-$version"-linux-ppc64le" linux ppc64le
# Build $name-$version"-linux-s390x" linux s390x

# other
# $name-$version
# Build $name-$version"-solaris-amd64" solaris amd64
Build $name-$version"-freebsd-386" freebsd 386
# Build $name-$version"-freebsd-amd64" freebsd amd64
# Build $name-$version"-freebsd-arm" freebsd arm
# Build $name-$version"-netbsd-386" netbsd	386
# Build $name-$version"-netbsd-amd64" netbsd amd64
# Build $name-$version"-netbsd-arm" netbsd	arm
# Build $name-$version"-openbsd-386" openbsd 386
# Build $name-$version"-openbsd-amd64" openbsd	amd64
# Build $name-$version"-openbsd-arm" openbsd arm
# Build $name-$version"-plan9-386" plan9 386
# Build $name-$version"-plan9-amd64" plan9 amd64
# Build $name-$version"-plan9-arm" plan9 arm
# Build $name-$version"-nacl-386" nacl 386
# Build $name-$version"-nacl-amd64p32" nacl amd64p32
# Build $name-$version"-nacl-arm" nacl arm
# Build $name-$version"-dragonflybsd-amd64" dragonfly amd64
