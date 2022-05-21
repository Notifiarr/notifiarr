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

RUN mkdir -p $GOPATH/pkg/mod $GOPATH/bin $GOPATH/src /${BINARY}
COPY . /${BINARY}
WORKDIR /${BINARY}

RUN CGO_ENABLED=0 make clean ${BINARY}.${TARGETARCH}.${TARGETOS} && \
    curl -sSo /notifiarr.tgz https://raw.githubusercontent.com/Notifiarr/build-dependencies/main/notifiarr-docker-$TARGETARCH.tgz && \
    tar -zxf /notifiarr.tgz -C / && \
    ln -sf /lib/x86_64-linux-gnu/libncurses.so.6 /lib/x86_64-linux-gnu/libncurses.so.5

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

# Make sure we have an ssl cert chain and timezone data.
COPY --from=builder /etc/ssl /etc/ssl
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo

ENV TZ=UTC

# Notifiarr specific.
ENV PATH=/
ENV USER=root
ENV NOTIFIARR_IN_DOCKER=true
EXPOSE 5454
COPY --from=builder /MegaCli* /libstorelibir-2.so.14.07-0 /smartctl /
# All of this is for megacli, and probably doesn't work on i386.
COPY --from=builder /lib64/ld-linux-x86-64.so.2 /lib64/ld-linux-x86-64.so.2 /lib64/
COPY --from=builder /usr/lib/x86_64-linux-gnu/libstdc++.so.6 /usr/lib/x86_64-linux-gnu/libstdc++.so.6.0.25 \
  /lib/x86_64-linux-gnu/ld-2.28.so /lib/x86_64-linux-gnu/libpthread.so.0 /lib/x86_64-linux-gnu/libpthread-2.28.so \
  /lib/x86_64-linux-gnu/libm.so.6 /lib/x86_64-linux-gnu/libm-2.28.so /lib/x86_64-linux-gnu/libgcc_s.so.1 /lib/x86_64-linux-gnu/libdl.so.2 \
  /lib/x86_64-linux-gnu/libdl-2.28.so /lib/x86_64-linux-gnu/libc.so.6 /lib/x86_64-linux-gnu/libc-2.28.so /lib/x86_64-linux-gnu/libncurses.so.5 \
  /lib/x86_64-linux-gnu/libncurses.so.6 /lib/x86_64-linux-gnu/libncurses.so.6.1 /lib/x86_64-linux-gnu/libtinfo.so.6 /lib/x86_64-linux-gnu/

ENTRYPOINT [ "/image" ]
