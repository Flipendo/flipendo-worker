FROM debian:jessie
MAINTAINER Guillaume "Le G" Legros
RUN apt-get update && apt-get install -y gcc make golang git yasm openssl libssl-dev libx264-dev
RUN git clone https://github.com/FFmpeg/FFmpeg.git && cd FFmpeg && ./configure --enable-libx264 --enable-gpl --enable-openssl --enable-nonfree && make && make install
RUN mkdir /GO
ENV LD_LIBRARY_PATH $LD_LIBRARY_PATH:/usr/local/lib
ENV GOPATH /GO
COPY . /GO/src/github.com/flipendo/flipendo-worker
WORKDIR /GO/src/github.com/flipendo/flipendo-worker
RUN go get ./...
RUN go install
CMD ["/GO/bin/flipendo-worker"]
