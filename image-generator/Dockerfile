FROM alpine
# needed to be able to connect to https sources
RUN  apk add --update ca-certificates ffmpeg && rm -rf /var/cache/apk/*
RUN  apk add --update bash && rm -rf /var/cache/apk/*
COPY resources .
COPY ./image-generator image-generator
ENTRYPOINT ["/image-generator"]