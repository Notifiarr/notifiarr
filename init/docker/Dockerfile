#
# This is part of Application Builder.
# https://github.com/golift/application-builder
#

FROM golang:1-bullseye as builder

# Build speed magic! 
# https://www.docker.com/blog/faster-multi-platform-builds-dockerfile-cross-compilation-guide/
WORKDIR /src
COPY main.go settings.sh go.mod go.sum ./
RUN go mod download
RUN go install github.com/swaggo/swag/cmd/swag@v1.8.7
RUN go install github.com/kevinburke/go-bindata/...@latest
COPY pkg pkg
RUN go generate ./...

RUN apt update
RUN apt install -y libncurses5 libstdc++6 libtinfo5 apt-utils ca-certificates

ARG TARGETARCH

RUN curl -sSo /notifiarr.tgz \
 https://raw.githubusercontent.com/Notifiarr/build-dependencies/main/notifiarr-docker-$TARGETARCH.tgz \
 && tar -zxf /notifiarr.tgz -C /

# For megacli. All the *'s are to deal with multiarch. :(
RUN mkdir -p /tmp/lib_link /tmp$(ls -d /lib/*-linux-gnu*) && cp /usr/lib/*-linux-gnu*/libstdc++.so* \
 /lib/*-linux-gnu*/ld-2.*.so /lib/*-linux-gnu*/libpthread.so.0 /lib/*-linux-gnu*/libpthread-2.*.so \
 /lib/*-linux-gnu*/libm.so.6 /lib/*-linux-gnu*/libm-2.*.so /lib/*-linux-gnu*/libgcc_s.so.1 \
 /lib/*-linux-gnu*/libdl.so.2 /lib/*-linux-gnu*/libdl-2.*.so /lib/*-linux-gnu*/libc.so.6 \
 /lib/*-linux-gnu*/libc-2.*.so /lib/*-linux-gnu*/libncurses.so.5 \
 /lib/*-linux-gnu*/libtinfo.so.5 /tmp$(ls -d /lib/*-linux-gnu*) && \
 ln -s /lib/*-linux-gnu*/ld-2.*.so /tmp/lib/ld-linux-x86-64.so.2 && \
 ln -s /lib/*-linux-gnu*/ld-2.*.so /tmp/lib/ld-linux-aarch64.so.1 && \
 ln -s /usr/lib /tmp/lib_link/lib64 && \
 ln -s /usr/lib /tmp/lib_link/lib

ARG TARGETOS
ARG BUILD_DATE
ARG BRANCH
ARG VERSION
ARG ITERATION
ARG COMMIT

# Build the app.
RUN GOFLAGS="-trimpath -mod=readonly -modcacherw" \
 GOOS=${TARGETOS} GOARCH=${TARGETARCH} go build -o /tmp/image -tags osusergo,netgo \
 -ldflags "-w -s -X \"golift.io/version.Branch=${BRANCH} (${COMMIT})\" \
 -X \"golift.io/version.BuildDate=${BUILD_DATE}\" \
 -X \"golift.io/version.BuildUser=docker\" \
 -X \"golift.io/version.Revision=${ITERATION}\" \
 -X \"golift.io/version.Version=${VERSION}\""

FROM scratch

# Make sure we have an ssl cert chain and timezone data.
COPY --from=builder /etc/ssl /etc/ssl
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo

COPY --from=builder /tmp/image /image

# For megacli.
COPY --from=builder /MegaCli* /libstorelibir-2.so.14.07-0 /smartctl /
COPY --from=builder /tmp/lib /usr/lib
COPY --from=builder /tmp/lib_link/ /

ENV TZ=UTC

# Notifiarr specific.
ENV PATH=/
ENV USER=root
ENV NOTIFIARR_IN_DOCKER=true

EXPOSE 5454
ENTRYPOINT [ "/image" ]

ARG TARGETOS
ARG TARGETARCH
ARG BUILD_DATE
ARG COMMIT
ARG VERSION
ARG ITERATION
ARG LICENSE=MIT
ARG SOURCE_URL=http://github.com/golift/application-builder
ARG DESC=application-builder
ARG VENDOR=golift
ARG AUTHOR=golift

# Build-time metadata as defined at https://github.com/opencontainers/image-spec/blob/master/annotations.md
LABEL org.opencontainers.image.created="${BUILD_DATE}" \
      org.opencontainers.image.title="Notifiarr" \
      org.opencontainers.image.documentation="${SOURCE_URL}/wiki/Docker" \
      org.opencontainers.image.description="${DESC}" \
      org.opencontainers.image.url="${SOURCE_URL}" \
      org.opencontainers.image.revision="${COMMIT}" \
      org.opencontainers.image.source="${SOURCE_URL}" \
      org.opencontainers.image.vendor="${VENDOR}" \
      org.opencontainers.image.authors="${AUTHOR}" \
      org.opencontainers.image.architecture="${TARGETOS} ${TARGETARCH}" \
      org.opencontainers.image.licenses="${LICENSE}" \
      org.opencontainers.image.version="${VERSION}-${ITERATION}"