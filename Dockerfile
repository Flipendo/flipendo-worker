FROM debian:jessie
MAINTAINER Guillaume "Le G" Legros
RUN apt-get update && apt-get install -y gcc make golang git yasm
RUN git clone git://git.videolan.org/x264.git && cd x264 && ./configure --enable-static --enable-shared && make && make install
RUN git clone https://github.com/FFmpeg/FFmpeg.git && cd FFmpeg && ./configure --enable-libx264 --enable-gpl && make && make install


