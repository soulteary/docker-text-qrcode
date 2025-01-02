FROM FROM alpine:3.20
RUN apk update && \
    apk add --no-cache build-base libqrencode-dev
COPY . .
RUN make clean && \
    make -j && \
    make install

CMD ["qr"]
