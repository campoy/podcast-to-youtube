FROM alpine
# needed to be able to connect to https sources
RUN  apk add --update ca-certificates ffmpeg && rm -rf /var/cache/apk/*
COPY resources .
COPY ./media-generator media-generator
ENTRYPOINT ["/media-generator"]