FROM alpine:latest AS builder

RUN apk update && apk add --no-cache \
    build-base \
    pkgconf \
    curl \
    autoconf \
    automake \
    libtool \
    bison \
    openssl-dev

# THe place for libraries
ENV TARGETDIR=/usr/local

WORKDIR /tmp

# Build libevent
RUN curl -LO https://github.com/libevent/libevent/releases/download/release-2.1.12-stable/libevent-2.1.12-stable.tar.gz \
    && tar -zxvf libevent-2.1.12-stable.tar.gz \
    && cd libevent-2.1.12-stable \
    && ./configure --prefix=$TARGETDIR --enable-shared && make && make install

# Build ncurses
RUN curl -LO https://ftp.gnu.org/pub/gnu/ncurses/ncurses-6.4.tar.gz \
    && tar zxvf ncurses-6.4.tar.gz \
    && cd ncurses-6.4 \
    && ./configure --prefix=$TARGETDIR --with-default-terminfo-dir=/usr/share/terminfo --with-shared --with-terminfo-dirs="/etc/terminfo:/lib/terminfo:/usr/share/terminfo" --enable-pc-files --with-pkg-config-libdir=$TARGETDIR/lib/pkgconfig \
    && make && make install

# Download and build tmux
RUN curl -LO https://github.com/tmux/tmux/releases/download/3.4/tmux-3.4.tar.gz \
    && tar zxvf tmux-3.4.tar.gz \
    && cd tmux-3.4 \
    && PKG_CONFIG_PATH=$TARGETDIR/lib/pkgconfig ./configure --enable-static --prefix=$TARGETDIR && make && make install

FROM scratch AS export-stage
COPY --from=builder /usr/local/bin/tmux /bin/tmux
