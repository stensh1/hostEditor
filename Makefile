.PHONY:
.SILENT:

## Generate grpc
# Installing brew manager
getBrew:
	/bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)"; \
	echo 'eval "$(/home/linuxbrew/.linuxbrew/bin/brew shellenv)"' >> ~/.bash_profile; \
    eval "$(/home/linuxbrew/.linuxbrew/bin/brew shellenv)"
# Installing buf through the brew
getBuf: getBrew
	brew install bufbuild/buf/buf
# Generating service code
genBuf:
	cd api/proto/; buf generate


## Server
# Build server
sbuild:
	go build -o ./.bin/server ./cmd/server/main.go
# Run server
srun:
	./.bin/server


## Client
# Build client
cbuild:
	go build -o ./.bin/client ./cmd/client/main.go
# Run client
crun:
	./.bin/client