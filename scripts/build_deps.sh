#!/bin/bash

set -e

# 设置变量
PROJECT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" &> /dev/null && pwd)/.."
INSTALL_DIR="$PROJECT_DIR/.build_cache"

CGO_DIR="$PROJECT_DIR/font"
THIRD_PARTY_DIR="$PROJECT_DIR/third_party"

INCLUDE_DIR="$CGO_DIR/include"
LIB_DIR="$CGO_DIR/libs"

# 初始化
rm -rf "$INSTALL_DIR"
rm -rf "$LIB_DIR"
mkdir -p "$LIB_DIR"

echo "Building C dependencies..."

# 构建 zlib
echo "Building zlib..."
cd "$THIRD_PARTY_DIR/zlib"
make clean || true
./configure --prefix="$INSTALL_DIR" --static
make
make install

# 构建 libpng (依赖 zlib)
echo "Building libpng..."
cd "$THIRD_PARTY_DIR/libpng"
make clean || true
./configure --prefix="$INSTALL_DIR" --enable-static --disable-shared \
    --with-zlib-prefix="$INSTALL_DIR"
make
make install

# 构建 bzip2
echo "Building bzip2..."
cd "$THIRD_PARTY_DIR/bzip2"
rm -rf build
mkdir -p build && cd build
cmake -DCMAKE_INSTALL_PREFIX="$INSTALL_DIR" \
      -DCMAKE_BUILD_TYPE=Release \
      -DENABLE_STATIC_LIB=ON \
      -DENABLE_SHARED_LIB=OFF \
      -DENABLE_LIB_ONLY=ON \
      -DENABLE_APP=OFF \
      -DENABLE_EXAMPLES=OFF \
      -DENABLE_DOCS=OFF \
      ..
make
make install

# 构建 brotli
echo "Building brotli..."
cd "$THIRD_PARTY_DIR/brotli"
mkdir -p build && cd build
cmake -DCMAKE_INSTALL_PREFIX="$INSTALL_DIR" \
      -DCMAKE_BUILD_TYPE=Release \
      -DBUILD_SHARED_LIBS=OFF \
      ..
make
make install

echo "Base dependencies built successfully!"
echo "Now checking if we have the required files..."

# 将库文件从lib目录复制到libs目录
echo "Copying libraries from lib to libs directory..."
cp -f "$INSTALL_DIR/lib"/*.a "$LIB_DIR/"

# 创建符号链接或重命名文件以保持一致的命名
if [ -f "$LIB_DIR/libbz2_static.a" ] && [ ! -f "$LIB_DIR/libbz2.a" ]; then
    cp "$LIB_DIR/libbz2_static.a" "$LIB_DIR/libbz2.a"
fi

# 检查构建结果
if [ -f "$LIB_DIR/libz.a" ]; then
    echo "✓ zlib built successfully"
else
    echo "✗ zlib not found" 
    exit 1
fi

if [ -f "$LIB_DIR/libpng16.a" ]; then
    echo "✓ libpng built successfully"
else
    echo "✗ libpng not found"
    exit 1
fi

if [ -f "$LIB_DIR/libbz2.a" ]; then
    echo "✓ bzip2 built successfully"
else
    echo "✗ bzip2 not found"
    exit 1
fi

if [ -f "$LIB_DIR/libbrotlicommon.a" ] && [ -f "$LIB_DIR/libbrotlidec.a" ] && [ -f "$LIB_DIR/libbrotlienc.a" ]; then
    echo "✓ brotli built successfully"
else
    echo "✗ brotli not found"
    exit 1
fi

echo ""
echo "Building FreeType..."
# 构建 FreeType (依赖 zlib, libpng, bzip2, brotli)
cd "$THIRD_PARTY_DIR/freetype"
make clean || true

# 设置PKG_CONFIG_PATH以便FreeType能找到依赖
export PKG_CONFIG_PATH="$LIB_DIR/pkgconfig:$PKG_CONFIG_PATH"

./autogen.sh
./configure --prefix="$INSTALL_DIR" \
    --enable-static --disable-shared \
    --with-zlib="$INSTALL_DIR" \
    --with-png="$INSTALL_DIR" \
    --with-bzip2="$INSTALL_DIR" \
    --with-brotli="$INSTALL_DIR"
make
make install
cp "$INSTALL_DIR/lib/libfreetype.a" "$CGO_DIR/libs"

echo ""
echo "Building HarfBuzz..."
# 构建 HarfBuzz (依赖 FreeType)
cd "$THIRD_PARTY_DIR/harfbuzz"
rm -rf build
mkdir -p build && cd build
meson setup --prefix="$INSTALL_DIR" \
    --default-library=static \
    --buildtype=release \
    -Dfreetype=enabled \
    -Dglib=disabled \
    -Dgobject=disabled \
    -Dcairo=disabled \
    -Dicu=disabled \
    -Dgraphite=disabled \
    -Dintrospection=disabled \
    -Ddocs=disabled \
    -Dtests=disabled \
    -Dbenchmark=disabled \
    -Dexperimental_api=true \
    ..
ninja
ninja install
cp "$INSTALL_DIR/lib/libharfbuzz.a" "$CGO_DIR/libs"
cp "$INSTALL_DIR/lib/libharfbuzz-subset.a" "$CGO_DIR/libs"

cp -r  "$INSTALL_DIR/include" "$CGO_DIR"

echo ""
echo "All dependencies built successfully!"
echo "Header files are in: $INCLUDE_DIR"
echo "Library files are in: $LIB_DIR"
