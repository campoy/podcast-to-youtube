FROM alpine
RUN  apk add --update ca-certificates ffmpeg && rm -rf /var/cache/apk/*
COPY resources .
COPY client_secrets.json .
COPY ./podcast2youtube .
ENTRYPOINT ["/podcast2youtube"]