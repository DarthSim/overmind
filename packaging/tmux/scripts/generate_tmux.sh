#!/bin/bash

if [ -z "$TMUX_INSTALL_DIR" ]; then
  TMUX_INSTALL_DIR="/usr/local"
fi

TMUX_LIB_DIR="$TMUX_INSTALL_DIR/lib"

OS=$(uname)
echo "Detected OS: $OS"

mkdir -p "$TMUX_INSTALL_DIR/bin"
mkdir -p "$TMUX_LIB_DIR"
LDFLAGS="-L$TMUX_LIB_DIR"
CPPFLAGS="-I$TMUX_LIB_DIR/include"
LIBS="-lresolv"
TMUX_FLAGS="--enable-static"

if [ "$OS" == "Darwin" ]; then
    TMUX_FLAGS="--enable-utf8proc"
fi

TARGETDIR="$TMUX_INSTALL_DIR"

# Build libevent
curl -LO https://github.com/libevent/libevent/releases/download/release-2.1.12-stable/libevent-2.1.12-stable.tar.gz \
  && tar -zxvf libevent-2.1.12-stable.tar.gz \
  && cd libevent-2.1.12-stable \
  && LDFLAGS="$LDFLAGS" CPPFLAGS="$CPPFLAGS" ./configure --prefix=$TARGETDIR --enable-shared && make && make install \
  && cd ..

# Build ncurses
curl -LO https://ftp.gnu.org/pub/gnu/ncurses/ncurses-5.9.tar.gz \
  && tar zxvf ncurses-5.9.tar.gz \
  && cd ncurses-5.9 \
  && LDFLAGS="$LDFLAGS" CPPFLAGS="$CPPFLAGS" ./configure --prefix=$TARGETDIR --with-shared --with-termlib --enable-pc-files && make && make install \
  && cd ..

# utf8proc for macos
curl -LO https://github.com/JuliaStrings/utf8proc/releases/download/v2.9.0/utf8proc-2.9.0.tar.gz \
  && tar zxvf utf8proc-2.9.0.tar.gz \
  && cd utf8proc-2.9.0 \
  && LDFLAGS="$LDFLAGS" CPPFLAGS="$CPPFLAGS" make prefix=$TARGETDIR && make install prefix=$TARGETDIR \
  && cd ..

# Download and build tmux
curl -LO https://github.com/tmux/tmux/releases/download/3.4/tmux-3.4.tar.gz \
  && tar zxvf tmux-3.4.tar.gz \
  && cd tmux-3.4 \
  && PKG_CONFIG_PATH="$TMUX_LIB_DIR/pkgconfig" LDFLAGS="$LDFLAGS" CPPFLAGS="$CPPFLAGS" LIBS="-lresolv -lutf8proc" ./configure $TMUX_FLAGS --prefix=$TARGETDIR && make && make install \
  && cd ..

if [ "$OS" == "Linux" ]; then
    find "$TARGETDIR/lib" -name "libutf8proc*.so*" -exec cp {} "$TMUX_LIB_DIR" \;
fi

# Copy all libraries to TMUX_LIB_DIR and refresh dependencies` paths for MacOS using install_name_tool
if [ "$OS" == "Darwin" ]; then
    find "$TARGETDIR/lib" -name "*.dylib" -exec cp {} "$TMUX_LIB_DIR" \;

    for lib in "$TMUX_LIB_DIR"/*.dylib; do
        libname=$(basename "$lib")
        install_name_tool -change "$TARGETDIR/lib/$libname" "@executable_path/../lib/$libname" "$TARGETDIR/bin/tmux"
    done
fi

mv "$TARGETDIR/bin/tmux" "$TMUX_INSTALL_DIR/bin/tmux"

echo "tmux has been installed to $TMUX_INSTALL_DIR/bin/tmux with dependencies in $TMUX_LIB_DIR"
