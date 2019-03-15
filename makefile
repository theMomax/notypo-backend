all: help

# from https://gist.github.com/prwhite/8168133
help: ## This help dialog.
	@IFS=$$'\n' ; \
	help_lines=(`fgrep -h "##" $(MAKEFILE_LIST) | fgrep -v fgrep | sed -e 's/\\$$//'`); \
	for help_line in $${help_lines[@]}; do \
		IFS=$$'#' ; \
		help_split=($$help_line) ; \
		help_command=`echo $${help_split[0]} | sed -e 's/^ *//' -e 's/ *$$//'` ; \
		help_info=`echo $${help_split[2]} | sed -e 's/^ *//' -e 's/ *$$//'` ; \
		printf "%-30s %s\n" $$help_command $$help_info ; \
	done

setup: gosetup configsetup ## Prepares the environment for running the server. This includes:
##					 - Downloading the go dependencies
##					 - Copying the default.config.ini to the default config_path (config.ini)

gosetup: ## Runs go mod download
	go mod download

configsetup: ## Copies the default.config.ini to the default config_path (config.ini).
	test -f config.ini || cp default.config.ini config.ini

test: ## Runs all package-tests.
	go test ./...

build: version = development
build: git_commit=$(shell git rev-list -1 HEAD)
build: config_path = config.ini
build: build_time = $(shell date "+%Y-%m-%d %H:%M:%S")


build: ## Builds the server. Make sure to set your GOOS and GOARCH to match your build target.
##				USAGE: make build [variable=value]
##				VARIABLES (DEFAULT_VALUE):
##				 	- version (develop)
##					- config_path (config.ini)
	go build -ldflags "-X github.com/theMomax/notypo-backend/config.GitCommit=$(git_commit) -X github.com/theMomax/notypo-backend/config.Version=$(version) -X 'github.com/theMomax/notypo-backend/config.ConfigPath=$(config_path)' -X 'github.com/theMomax/notypo-backend/config.BuildTime=$(build_time)'"

docs: ## Opens a browser window containing the api's swagger documentation. This command requires go-swagger.
	swagger serve swagger.yml
	#docker run -p 8080:80 -e SPEC_URL=file://$(PWD)/swagger.yml redocly/redoc