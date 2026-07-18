FROM ubuntu:24.04

RUN apt-get update && apt-get install -y \
    git \
    build-essential \
    autoconf \
    automake \
    libtool \
    pkg-config \
    gengetopt \
    libmicrohttpd-dev \
    libjansson-dev \
    libssl-dev \
    libsrtp2-dev \
    libsofia-sip-ua-dev \
    libglib2.0-dev \
    libopus-dev \
    libogg-dev \
    libcurl4-openssl-dev \
    liblua5.4-dev \
    libconfig-dev \
    libnice-dev \
    && rm -rf /var/lib/apt/lists/*

RUN git clone https://github.com/meetecho/janus-gateway.git /src/janus \
    && cd /src/janus \
    && git checkout v1.4.1 \
    && sh autogen.sh \
    && ./configure \
    && make -j$(nproc) \
    && make install \
    && make configs

EXPOSE 8088 8188 10000-10200/udp

CMD ["janus"]