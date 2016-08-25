build: binary docker clean
binary:
	GOOS=linux go build -i
docker:
	docker build -t $(USER)/podcast2youtube .
clean:
	rm -f podcast2youtube

run: build
	docker run --rm -it $(USER)/podcast2youtube
