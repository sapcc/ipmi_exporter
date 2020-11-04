FROM golang:1.14.0 AS build-env
WORKDIR /src
ADD . /src
RUN go get -d
RUN CGO_ENABLED=0 GOOS=linux go build -a -tags netgo -ldflags '-w' -o ipmi_exporter

FROM ubuntu:20.04
LABEL maintainer "sapcc <stefan.hipfel@sap.com>"
LABEL source_repository="https://github.com/sapcc/ipmi-exporter"
RUN apt-get update && apt-get install -y \
    wget \
    build-essential \
    libgcrypt20-dev

RUN wget http://ftp.gnu.org/gnu/freeipmi/freeipmi-1.6.6.tar.gz && tar xzvf freeipmi-1.6.6.tar.gz
WORKDIR /freeipmi-1.6.6
RUN ./configure && make && make install && ldconfig
WORKDIR /app
COPY --from=build-env /src/ipmi_exporter /app/
COPY ipmi.yml /app
EXPOSE 9290 
CMD ["./ipmi_exporter"]
