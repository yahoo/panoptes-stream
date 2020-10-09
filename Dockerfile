# build 
FROM golang:1.15.0 as builder
WORKDIR /go/src/
RUN mkdir -p github.com/yahoo/panoptes-stream
ADD . github.com/yahoo/panoptes-stream
WORKDIR /go/src/github.com/yahoo/panoptes-stream/panoptes
RUN CGO_ENABLED=0 go build .
WORKDIR /go/src/github.com/yahoo/panoptes-stream/telemetry/simulator/
RUN CGO_ENABLED=0 go build . 

# run 
FROM alpine:latest 
COPY --from=builder /go/src/github.com/yahoo/panoptes-stream/panoptes/panoptes /usr/bin/
COPY --from=builder /go/src/github.com/yahoo/panoptes-stream/telemetry/simulator/simulator /usr/bin/
EXPOSE 8081
ENTRYPOINT ["/usr/bin/panoptes"]