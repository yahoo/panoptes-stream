# build 
FROM golang:alpine as builder

WORKDIR /go/src/

RUN mkdir -p github.com/yahoo/panoptes-stream
COPY . github.com/yahoo/panoptes-stream

WORKDIR /go/src/github.com/yahoo/panoptes-stream/panoptes

RUN CGO_ENABLED=0 go build -ldflags="-X 'github.com/yahoo/panoptes-stream/config.version=v0.1.0'"

WORKDIR /go/src/github.com/yahoo/panoptes-stream/telemetry/simulator/

RUN CGO_ENABLED=0 go build 

# run 
FROM alpine 

COPY --from=builder /go/src/github.com/yahoo/panoptes-stream/panoptes/panoptes /usr/bin/
COPY --from=builder /go/src/github.com/yahoo/panoptes-stream/telemetry/simulator/simulator /usr/bin/

EXPOSE 8081

ENTRYPOINT ["panoptes"]