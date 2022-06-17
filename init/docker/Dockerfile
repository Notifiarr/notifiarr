#
# This is part of Application Builder.
# https://github.com/golift/application-builder
#

ARG BUILD_DATE=0
ARG COMMIT=0
ARG VERSION=unknown
ARG BINARY=application-builder

FROM golang:1-buster as builder
ARG TARGETOS
ARG BINARY
ARG TARGETARCH
ARG BUILD_FLAGS
RUN mkdir -p $GOPATH/pkg/mod $GOPATH/bin $GOPATH/src /${BINARY}
COPY . /${BINARY}
WORKDIR /${BINARY}

# For megacli. All the *'s are to deal with multiarch. :(
RUN apt update && apt install -y libncurses5 libstdc++6 libtinfo5 && \
    curl -sSo /notifiarr.tgz https://raw.githubusercontent.com/Notifiarr/build-dependencies/main/notifiarr-docker-$TARGETARCH.tgz && \
    tar -zxf /notifiarr.tgz -C / && \
    mkdir -p /tmp/lib_link /tmp$(ls -d /lib/*-linux-gnu*) && cp /usr/lib/*-linux-gnu*/libstdc++.so* \
    /lib/*-linux-gnu*/ld-2.*.so /lib/*-linux-gnu*/libpthread.so.0 /lib/*-linux-gnu*/libpthread-2.*.so \
    /lib/*-linux-gnu*/libm.so.6 /lib/*-linux-gnu*/libm-2.*.so /lib/*-linux-gnu*/libgcc_s.so.1 \
    /lib/*-linux-gnu*/libdl.so.2 /lib/*-linux-gnu*/libdl-2.*.so /lib/*-linux-gnu*/libc.so.6 \
    /lib/*-linux-gnu*/libc-2.*.so /lib/*-linux-gnu*/libncurses.so.5 \
    /lib/*-linux-gnu*/libtinfo.so.5 /tmp$(ls -d /lib/*-linux-gnu*) && \
    ln -s /lib/*-linux-gnu*/ld-2.*.so /tmp/lib/ld-linux-x86-64.so.2 && \
    ln -s /lib/*-linux-gnu*/ld-2.*.so /tmp/lib/ld-linux-aarch64.so.1 && \
    ln -s /usr/lib /tmp/lib_link/lib64 && \
    ln -s /usr/lib /tmp/lib_link/lib

# Build the app.
RUN CGO_ENABLED=1 make clean ${BINARY}.${TARGETARCH}.${TARGETOS}

FROM scratch
ARG TARGETOS
ARG TARGETARCH
ARG BUILD_DATE
ARG COMMIT
ARG VERSION
ARG LICENSE=MIT
ARG BINARY
ARG SOURCE_URL=http://github.com/golift/application-builder
ARG DESC=application-builder
ARG VENDOR=golift
ARG AUTHOR=golift
ARG CONFIG_FILE=config.conf

# Build-time metadata as defined at https://github.com/opencontainers/image-spec/blob/master/annotations.md
LABEL org.opencontainers.image.created="${BUILD_DATE}" \
      org.opencontainers.image.title="${BINARY}" \
      org.opencontainers.image.documentation="${SOURCE_URL}/wiki/Docker" \
      org.opencontainers.image.description="${DESC}" \
      org.opencontainers.image.url="${SOURCE_URL}" \
      org.opencontainers.image.revision="${COMMIT}" \
      org.opencontainers.image.source="${SOURCE_URL}" \
      org.opencontainers.image.vendor="${VENDOR}" \
      org.opencontainers.image.authors="${AUTHOR}" \
      org.opencontainers.image.architecture="${TARGETOS} ${TARGETARCH}" \
      org.opencontainers.image.licenses="${LICENSE}" \
      org.opencontainers.image.version="${VERSION}"

COPY --from=builder /${BINARY}/${BINARY}.${TARGETARCH}.${TARGETOS} /image

# For megacli.
COPY --from=builder /MegaCli* /libstorelibir-2.so.14.07-0 /smartctl /
COPY --from=builder /tmp/lib /usr/lib
COPY --from=builder /tmp/lib_link/ /

# Make sure we have an ssl cert chain and timezone data.
COPY --from=builder /etc/ssl /etc/ssl
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo

ENV TZ=UTC

# Notifiarr specific.
ENV PATH=/
ENV USER=root
ENV NOTIFIARR_IN_DOCKER=true

EXPOSE 5454
ENTRYPOINT [ "/image" ]
