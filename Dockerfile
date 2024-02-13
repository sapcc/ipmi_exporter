FROM keppel.eu-de-1.cloud.sap/ccloud-dockerhub-mirror/library/golang:1.20.14-bullseye as builder
WORKDIR /go/src/github.com/prometheus-community/ipmi_exporter
RUN apt-get update && apt-get install -y make git
COPY . /src
RUN make -C /src install PREFIX=/build GO_BUILDFLAGS='-mod vendor'

FROM keppel.eu-de-1.cloud.sap/ccloud-dockerhub-mirror/library/debian:bullseye-slim
LABEL maintainer="The Prometheus Authors <prometheus-developers@googlegroups.com>"
LABEL source_repository="https://github.com/sapcc/ipmi_exporter"

RUN apt-get update && apt-get install -y freeipmi curl

WORKDIR /
RUN curl -Lo /bin/dumb-init https://github.com/Yelp/dumb-init/releases/download/v1.2.2/dumb-init_1.2.2_amd64 \
	&& chmod +x /bin/dumb-init \
	&& dumb-init -V

COPY --from=builder /build/ /usr/

EXPOSE      9290
USER        nobody
ENTRYPOINT ["dumb-init", "--"]
CMD ["/usr/bin/ipmi_exporter"]
