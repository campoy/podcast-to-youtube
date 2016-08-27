FROM alpine
RUN  apk add --update ca-certificates ffmpeg && rm -rf /var/cache/apk/*
COPY resources resources
COPY client_secrets.json .
COPY ./cmd .
ENTRYPOINT ["/cmd"]