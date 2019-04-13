GITHUB_OAUTH_CLIENT_ID = 39c483e563cd5cedf7c1
GITHUB_OAUTH_CLIENT_SECRET = 024b16270452504c35f541aca4bf78781cd06db9
APP_NAME = $(shell basename $(GOGEN_EMBED))
GOVVV = $(shell govvv -flags)

ifeq ($(OS),Windows_NT)
	dockercmd := docker run -e TERM -e HOME=/go/src/github.com/coccyx/gogen --rm -it -v $(CURDIR):/go/src/github.com/coccyx/gogen -v $(HOME)/.ssh:/root/.ssh clintsharp/gogen bash
else
	cd := $(shell pwd)
	dockercmd := docker run --rm -it -v $(cd):/go/src/github.com/coccyx/gogen clintsharp/gogen bash
endif

all: install

build:
	godep go build -ldflags "-X github.com/coccyx/gogen/internal.gitHubClientID=$(GITHUB_OAUTH_CLIENT_ID) -X github.com/coccyx/gogen/internal.gitHubClientSecret=$(GITHUB_OAUTH_CLIENT_SECRET) $(GOVVV)"

install:
	godep go install -ldflags "-X github.com/coccyx/gogen/internal.gitHubClientID=$(GITHUB_OAUTH_CLIENT_ID) -X github.com/coccyx/gogen/internal.gitHubClientSecret=$(GITHUB_OAUTH_CLIENT_SECRET) $(GOVVV)"

test:
	godep go test -v ./...

docker:
	$(dockercmd)

splunkapp:
	tar cfz splunk_app_gogen.spl splunk_app_gogen

embed:
	@if [ -z $$GOGEN_EMBED ]; then \
		echo "Set GOGEN_EMBED to the directory of your Splunk app to embed into."; \
		exit 1; \
	fi

	mkdir -p $$GOGEN_EMBED/bin
	cp splunk_app_gogen/bin/gogen.py $$GOGEN_EMBED/bin/$(APP_NAME)_gogen.py
	sed -i 's%<title>Gogen</title>%<title>$(APP_NAME) Gogen</title>%' $$GOGEN_EMBED/bin/$(APP_NAME)_gogen.py
	sed -i 's%<description>Generate data using Gogen</description>%<description>Generate data using $(APP_NAME) Gogen</description>%' $$GOGEN_EMBED/bin/$(APP_NAME)_gogen.py
	mkdir -p $$GOGEN_EMBED/default/data/ui/manager
	cp splunk_app_gogen/default/data/ui/manager/gogen_manager.xml $$GOGEN_EMBED/default/data/ui/manager/$(APP_NAME)_gogen_manager.xml
	sed -i 's%name="data/inputs/gogen"%name="data/inputs/$(APP_NAME)_gogen"%' $$GOGEN_EMBED/default/data/ui/manager/$(APP_NAME)_gogen_manager.xml 
	sed -i 's%<header>Gogen</header>%<header>$(APP_NAME) Gogen</header>%' $$GOGEN_EMBED/default/data/ui/manager/$(APP_NAME)_gogen_manager.xml
	sed -i 's%<name>Gogen</name>%<name>$(APP_NAME) Gogen</name>%' $$GOGEN_EMBED/default/data/ui/manager/$(APP_NAME)_gogen_manager.xml
	cat splunk_app_gogen/README/inputs.conf.spec >> $$GOGEN_EMBED/README/inputs.conf.spec
	sed -i 's%\[gogen://<name>\]%[$(APP_NAME)_gogen://<name>]%' $$GOGEN_EMBED/README/inputs.conf.spec