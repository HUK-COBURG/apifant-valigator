ifndef VALIDATENGINE_IMAGE
VALIDATENGINE_IMAGE := validatengine:latest
endif

.PHONY: build
build:
	docker build -t $(VALIDATENGINE_IMAGE) .

