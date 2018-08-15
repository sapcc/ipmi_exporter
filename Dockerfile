# This Dockerfile builds an image for a client_golang example.
#
# Use as (from the root for the client_golang repository):
#    docker build -f examples/$name/Dockerfile -t prometheus/golang-example-$name .

# Builder image, where we build the example.
FROM golang:1.9.0 AS build-env
WORKDIR /src
ADD . /src
RUN go get -d
RUN CGO_ENABLED=0 GOOS=linux go build -a -tags netgo -ldflags '-w' -o ipmi_exporter

# Final image.
FROM ubuntu:14.04
LABEL maintainer "sapcc <stefan.hipfel@sap.com>"
RUN sudo apt-get update && sudo apt-get install -y \
    wget \
    build-essential \
    libgcrypt11-dev

RUN wget http://ftp.gnu.org/gnu/freeipmi/freeipmi-1.6.2.tar.gz && tar xzvf freeipmi-1.6.2.tar.gz
WORKDIR /freeipmi-1.6.2
RUN ./configure && make && sudo make install && sudo ldconfig
WORKDIR /app
COPY --from=build-env /src/ipmi_exporter /app/
COPY ipmi.yml /app
EXPOSE 9290 
ENTRYPOINT ["./ipmi_exporter"]