build: binary docker clean
binary:
	GOOS=linux go build -i -o cmd
docker:
	docker build -t $(USER)/podcast2youtube .
clean:
	rm -f cmd

run: build
	docker run --rm -it $(USER)/podcast2youtube
