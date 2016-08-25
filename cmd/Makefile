build: binary docker clean
binary:
	GOOS=linux go build -i
docker:
	docker build -t gcr.io/podcast-to-youtube/cmd .
clean:
	rm -f cmd

run:
	docker run --rm -it -v \
		"$(HOME)/.config/gcloud/application_default_credentials.json":"/root/.config/gcloud/application_default_credentials.json" \
		gcr.io/podcast-to-youtube/cmd
bash:
	docker run --rm -it -v \
		"$(HOME)/.config/gcloud/application_default_credentials.json":"/root/.config/gcloud/application_default_credentials.json" \
		-v "$(PWD):/data" -w "/data" \
		--entrypoint /bin/bash \
		gcr.io/podcast-to-youtube/cmd
