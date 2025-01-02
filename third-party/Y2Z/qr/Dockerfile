FROM alpine:3.20 AS builder
RUN apk update && \
    apk add --no-cache build-base libqrencode-dev
COPY . .
RUN make clean && \
    make -j && \
    make install

FROM alpine:3.20
COPY --from=builder /usr/lib/libqrencode.so.4.1.1 /usr/lib/libqrencode.so.4.1.1
RUN ln -s /usr/lib/libqrencode.so.4.1.1 /usr/lib/libqrencode.so.4
COPY --from=builder /usr/local/bin/qr /usr/local/bin/qr
CMD ["qr"]
