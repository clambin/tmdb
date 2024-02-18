FROM alpine

RUN apk update && \
    apk add --no-cache tzdata

WORKDIR /app

ARG TARGETOS
ARG TARGETARCH

COPY ./bld/tmdb-proxy-$TARGETOS-$TARGETARCH /app/tmdb-proxy

EXPOSE 8080
ENTRYPOINT ["/app/tmdb-proxy"]
CMD []
