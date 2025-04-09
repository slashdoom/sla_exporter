##
## Build
##

# pull official base image
FROM golang:alpine as build-env

# set work directory
WORKDIR /go/sla_exporter

# copy project from local
COPY . /go/sla_exporter

# get modules
RUN go mod download

# build sla_exporter binary
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -gcflags "all=-N -l" -o ./build/sla_exporter



##
## Deploy
##

# pull official base image
FROM golang:alpine

# set work directory
WORKDIR /go/sla_exporter

# copy binary from build-env container
COPY --from=build-env /go/sla_exporter/build/sla_exporter ./

# run binary
CMD ["./sla_exporter"]
