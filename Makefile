.PHONY: preview-isr release
preview-isr:
	@bash ./scripts/preview-isr.sh

release:
ifndef VERSION
	$(error VERSION is required, e.g. make release VERSION=v1.2.3 [PUSH=1])
endif
ifeq ($(PUSH),1)
	@bash ./scripts/release.sh $(VERSION) --push
else
	@bash ./scripts/release.sh $(VERSION)
endif
