PLUGIN_ID := com.akinlosotutech.onboardinghelper
PLUGIN_VERSION := $(shell jq -r .version plugin.json)
SERVER_DIR := server
DIST_DIR := dist
SERVER_DIST := $(SERVER_DIR)/dist
SERVER_BINARY := $(SERVER_DIST)/plugin-linux-amd64
GOCACHE := $(CURDIR)/.gocache
GOMODCACHE := $(CURDIR)/.gomodcache
GO := /usr/local/go/bin/go

.PHONY: build package clean

build:
	@mkdir -p $(SERVER_DIST)
	GOCACHE=$(GOCACHE) GOMODCACHE=$(GOMODCACHE) GOOS=linux GOARCH=amd64 $(GO) build -o $(SERVER_BINARY) ./$(SERVER_DIR)

package: build
	@rm -rf $(DIST_DIR)/$(PLUGIN_ID)
	@mkdir -p $(DIST_DIR)/$(PLUGIN_ID)/server/dist
	cp plugin.json $(DIST_DIR)/$(PLUGIN_ID)/
	cp -R $(SERVER_DIST)/. $(DIST_DIR)/$(PLUGIN_ID)/server/dist/
ifneq ("$(wildcard assets)","")
	cp -R assets $(DIST_DIR)/$(PLUGIN_ID)/
endif
	tar -czf $(DIST_DIR)/$(PLUGIN_ID)-$(PLUGIN_VERSION).tar.gz -C $(DIST_DIR)/$(PLUGIN_ID) .

clean:
	rm -rf $(DIST_DIR) $(SERVER_DIST)
