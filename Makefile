.PHONY: default
default: build start

build/spectral-package.zip:
	mkdir -p build
	cd spectral-package && zip -r ../build/spectral-package.zip .

.PHONY: build
build:
	docker compose build

.PHONY: start
start: build/spectral-package.zip build
	docker compose up -d

.PHONY: stop
stop:
	docker compose down -v