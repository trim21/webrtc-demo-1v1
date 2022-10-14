FROM golang:1.19
WORKDIR /app
ENV CGO_ENABLED=0

COPY . /src

RUN go build -o /rtc-signal-server
EXPOSE 8085
CMD [ "/rtc-signal-server" ]
