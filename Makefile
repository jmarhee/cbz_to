SHELL=/bin/bash
BINARY_NAME=watcher

all: build-linux-amd64 build-linux-arm64

build-linux-amd64:
	GOOS=linux GOARCH=amd64 go build -o $(BINARY_NAME)-linux-amd64 main.go

build-linux-arm64:
	GOOS=linux GOARCH=arm64 go build -o $(BINARY_NAME)-linux-arm64 main.go

clean:
	rm -f $(BINARY_NAME)-linux-amd64 $(BINARY_NAME)-linux-arm64

.SILENT:
tag-release:
	if [[ $(TAG) == v?.?.? ]]; then echo "Tagging $(TAG)"; elif [[ $(TAG) == v?.?.?? ]]; then echo "Tagging $(TAG)"; else echo "Bad Tag Format: $(TAG)"; exit 1; fi && git tag -a $(TAG) -m "Releasing $(TAG)" ; read -p "Push tag: $(TAG)? " push_tag ; if [ "${push_tag}"="yes" ]; then git push origin $(TAG); fi

.SILENT:
create-release:
	if [[ $(TAG) == v?.?.? ]]; then echo "Cutting release from $(TAG)"; elif [[ $(TAG) == v?.?.?? ]]; then echo "Cutting release from $(TAG)"; else echo "Bad Tag Format, cannot cut release: $(TAG)"; exit 1; fi && git tag -a $(TAG) -m "Releasing $(TAG)" ; read -p "Cut release from tag: $(TAG)? " push_tag ; if [ "${push_tag}"="yes" ]; then TAG=$(TAG) ./make-release.sh; fi
