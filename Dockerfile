# This Dockerfile builds an image for a client_golang example.
#
# Use as (from the root for the client_golang repository):
#    docker build -f examples/$name/Dockerfile -t prometheus/golang-example-$name .

# Builder image, where we build the example.
FROM golang:1.14 AS build-env
WORKDIR /src
ADD . /src
RUN go get -d
RUN CGO_ENABLED=0 GOOS=linux go build -a -tags netgo -ldflags '-w' -o ipmi_exporter

# Final image.
FROM ubuntu:18.04
LABEL maintainer "sapcc <stefan.hipfel@sap.com>"
LABEL source_repository="https://github.com/sapcc/ipmi-exporter"
RUN apt-get update && apt-get install -y \
    wget \
    build-essential \
    libgcrypt11-dev

RUN wget http://ftp.gnu.org/gnu/freeipmi/freeipmi-1.6.4.tar.gz && tar xzvf freeipmi-1.6.4.tar.gz
WORKDIR /freeipmi-1.6.4
RUN ./configure && make && make install && ldconfig
WORKDIR /app
COPY --from=build-env /src/ipmi_exporter /app/
COPY ipmi.yml /app
EXPOSE 9290 
ENTRYPOINT ["./ipmi_exporter"]
