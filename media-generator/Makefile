build: binary docker clean
binary:
	GOOS=linux go build
docker:
	docker build -t podcast-to-youtube/media-generator .
clean:
	rm -f media-generator

run:
	docker run --rm -it -v \
		"$(HOME)/.config/gcloud/application_default_credentials.json":"/root/.config/gcloud/application_default_credentials.json" \
		podcast-to-youtube/media-generator
bash:
	docker run --rm -it -v \
		"$(HOME)/.config/gcloud/application_default_credentials.json":"/root/.config/gcloud/application_default_credentials.json" \
		-v "$(PWD):/data" -w "/data" \
		--entrypoint /bin/bash \
		podcast-to-youtube/media-generator
