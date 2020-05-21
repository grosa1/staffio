.SILENT :
.PHONY : dep vet main clean dist package
DATE := `date '+%Y%m%d'`

WITH_ENV = env `cat .env 2>/dev/null | xargs`

ORIG:=liut7
NAME:=staffio
ROOF:=github.com/liut/$(NAME)
SOURCES=$(shell find cmd pkg templates -type f \( -name "*.go" ! -name "*_test.go" \) -print )
UIFILES=$(shell find fe/{css,scripts} -type f \( -name "*.styl" -o -name "*.js" \) -print )
STATICS=$(shell find htdocs -type f -print )
TAG:=`git describe --tags --always`
LDFLAGS:=-X $(ROOF)/pkg/settings.buildVersion=$(TAG)-$(DATE)
GO=$(shell which go)

main:
	echo "Building $(NAME)"
	$(GO) build -ldflags "$(LDFLAGS) -w" .

all: vet dist package

dep:
	$(GO) install golang.org/x/tools/go/analysis/passes/shadow/cmd/shadow
	$(GO) get github.com/liut/staticfiles

vet:
	echo "Checking ./pkg/..."
	$(GO) vet -all ./pkg/...

clean:
	echo "Cleaning dist"
	rm -rf dist fe/build
	rm -f $(NAME) $(NAME)-*
	rm -f .fe-build

dist/linux_amd64/$(NAME): $(SOURCES)
	echo "Building $(NAME) of linux"
	mkdir -p dist/linux_amd64 && cd dist/linux_amd64 && GOOS=linux GOARCH=amd64 $(GO) build -ldflags "$(LDFLAGS) -s -w" $(ROOF)

dist/darwin_amd64/$(NAME): $(SOURCES)
	echo "Building $(NAME) of darwin"
	mkdir -p dist/darwin_amd64 && cd dist/darwin_amd64 && GOOS=darwin GOARCH=amd64 $(GO) build -ldflags "$(LDFLAGS) -w" $(ROOF)

dist/windows_amd64/$(NAME): $(SOURCES)
	echo "Building $(NAME) of windows"
	mkdir -p dist/windows_amd64 && cd dist/windows_amd64 && GOOS=windows GOARCH=amd64 $(GO) build -ldflags "$(LDFLAGS) -s -w" $(ROOF)

dist: vet dist/linux_amd64/$(NAME) dist/darwin_amd64/$(NAME) dist/windows_amd64/$(NAME)


package: dist
	tar -cvJf $(NAME)-linux-amd64-$(TAG).tar.xz -C dist/linux_amd64 $(NAME)
	tar -cvJf $(NAME)-darwin-amd64-$(TAG).tar.xz -C dist/darwin_amd64 $(NAME)
	tar -cvJf $(NAME)-templates-$(TAG).tar.xz templates

fetch-exmail: # deprecated
	echo "Building $@"
	mkdir -p dist/linux_amd64 && GOOS=linux GOARCH=amd64 $(GO) build -ldflags "$(LDFLAGS)" -o dist/linux_amd64/$(NAME)-$@ $(ROOF)/cmd/$@
	mkdir -p dist/darwin_amd64 && GOOS=darwin GOARCH=amd64 $(GO) build -ldflags "$(LDFLAGS)" -o dist/darwin_amd64/$(NAME)-$@ $(ROOF)/cmd/$@
.PHONY: $@

wechat-work:
	echo "Building $@"
	mkdir -p dist/linux_amd64 && GOOS=linux GOARCH=amd64 $(GO) build -ldflags "$(LDFLAGS)" -o dist/linux_amd64/$(NAME)-$@ $(ROOF)/cmd/$@
	mkdir -p dist/darwin_amd64 && GOOS=darwin GOARCH=amd64 $(GO) build -ldflags "$(LDFLAGS)" -o dist/darwin_amd64/$(NAME)-$@ $(ROOF)/cmd/$@
.PHONY: $@

syncutil:
	echo "Building $@"
	mkdir -p dist/linux_amd64 && GOOS=linux GOARCH=amd64 $(GO) build -ldflags "$(LDFLAGS)" -o dist/linux_amd64/$(NAME)-$@ $(ROOF)/cmd/$@
	mkdir -p dist/darwin_amd64 && GOOS=darwin GOARCH=amd64 $(GO) build -ldflags "$(LDFLAGS)" -o dist/darwin_amd64/$(NAME)-$@ $(ROOF)/cmd/$@
.PHONY: $@

demo: # deprecated
	echo "Building $@"
	$(GO) build -ldflags "$(LDFLAGS)" $(ROOF)/cmd/$(NAME)-$@
.PHONY: $@

gen-key: # deprecated
	echo "Building $@"
	mkdir -p dist/linux_amd64 && GOOS=linux GOARCH=amd64 $(GO) build -ldflags "$(LDFLAGS)" -o dist/linux_amd64/$(NAME)-$@ $(ROOF)/cmd/gen-key
	mkdir -p dist/darwin_amd64 && GOOS=darwin GOARCH=amd64 $(GO) build -ldflags "$(LDFLAGS)" -o dist/darwin_amd64/$(NAME)-$@ $(ROOF)/cmd/gen-key
.PHONY: $@

fe-deps:
	npm install
.PHONY: $@

.fe-build: $(UIFILES)
	./node_modules/.bin/gulp clean build
	touch $@

fe-build: .fe-build

static:.fe-build $(STATICS)
	echo 'packing UI files into static'
	staticfiles --package static -o pkg/web/static/files.go ./htdocs
.PHONY: $@

gofmt:
	if [ -n "$(shell gofmt -l .)" ]; then \
		echo 1>&2 'The following files need to be formatted:'; \
		gofmt -l .; \
		exit 1; \
	fi

test-db: vet
	mkdir -p tests
	@$(WITH_ENV) go test -v -cover -coverprofile tests/cover_db.out ./pkg/backends
	@go tool cover -html=tests/cover_db.out -o tests/cover_db.out.html

test-ldap: vet
	mkdir -p tests
	@$(WITH_ENV) DEBUG=staffio:ldap go test -v -cover -coverprofile tests/cover_ldap.out ./pkg/backends/ldap
	@go tool cover -html=tests/cover_ldap.out -o tests/cover_ldap.out.html


docker-db-build:
	docker build --rm -t $(ORIG)/$(NAME)-db:$(TAG) database/
	docker tag $(ORIG)/$(NAME)-db:$(TAG) $(ORIG)/$(NAME)-db:latest

docker-db-save:
	docker save -o $(ORIG)_$(NAME)_db.tar $(ORIG)/$(NAME)-db:$(TAG) $(ORIG)/$(NAME)-db:latest && gzip -9f $(ORIG)_$(NAME)_db.tar

docker-auto-build:
	docker build --rm -t $(NAME) .

docker-local-build: dist/linux_amd64/$(NAME)
	echo "Building docker image"
	cp -rf templates entrypoint.sh dist/
	cp -rf Dockerfile.local dist/Dockerfile
	docker build --rm -t $(ORIG)/$(NAME):$(TAG) dist/
	docker tag $(ORIG)/$(NAME):$(TAG) $(ORIG)/$(NAME):latest
.PHONY: $@

docker-local-save:
	docker save -o $(ORIG)_$(NAME).tar $(ORIG)/$(NAME):$(TAG) $(ORIG)/$(NAME):latest && gzip -9f $(ORIG)_$(NAME).tar
.PHONY: $@

