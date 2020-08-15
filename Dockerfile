# build 
FROM golang:1.15.0 as builder
WORKDIR /go/src/
RUN mkdir -p git.vzbuilders.com/marshadrad/panoptes
ADD . git.vzbuilders.com/marshadrad/panoptes
WORKDIR /go/src/git.vzbuilders.com/marshadrad/panoptes/panoptes
RUN CGO_ENABLED=0 go build . 

# run 
FROM alpine:latest 
COPY --from=builder /go/src/git.vzbuilders.com/marshadrad/panoptes/panoptes/panoptes /usr/bin/
EXPOSE 8081
ENTRYPOINT ["/usr/bin/panoptes"]


# docker run -v $PWD/etc/:/etc panoptes -config /etc/config.yaml