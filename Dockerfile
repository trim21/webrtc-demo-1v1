FROM golang:1.19 as builder

ENV CGO_ENABLED=0
WORKDIR /app

COPY . /app/

RUN go build -o /app/rtc-signal-server

CMD [ "/app/rtc-signal-server" ]


FROM gcr.io/distroless/static

ENTRYPOINT ["/app/rtc-signal-server"]

COPY --from=builder /app/rtc-signal-server /app/rtc-signal-server
