FROM debian:jessie
MAINTAINER Guillaume "Le G" Legros
RUN apt-get update && apt-get install -y gcc make golang git yasm openssl libssl-dev libx264-dev
RUN git clone https://github.com/FFmpeg/FFmpeg.git && cd FFmpeg && ./configure --enable-libx264 --enable-gpl --enable-openssl --enable-nonfree && make && make install
RUN mkdir /go
ENV LD_LIBRARY_PATH $LD_LIBRARY_PATH:/usr/local/lib
ENV GOPATH /go
COPY . /go/src/github.com/flipendo/flipendo-worker
RUN cd /go/src/github.com/flipendo/flipendo-worker && go get ./... && go install && cd / && rm -rf /go/src /go/pkg
WORKDIR /flipendo
CMD ["/go/bin/flipendo-worker"]
