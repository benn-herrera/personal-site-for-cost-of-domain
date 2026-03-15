# Run from maint/: make site-tool
# Or directly:     make -C maint/site-tool [target]

BIN_DIR := ../bin

.PHONY: build clean dist nuke

# builds for current host platform only. product is .gitignored.
# cheap, infrequent op, no UTD check. primary use is development and unsupported platforms.
build:
	@mkdir -p $(BIN_DIR)
	go build -o $(BIN_DIR)/ .

clean:
	echo "cleaning local site-tool build"
	@rm -f $(BIN_DIR)/site-tool
	@rm -f $(BIN_DIR)/site-tool.exe


# builds all versions for checking into repo (LFS). infrequent, no UTD check.
dist:
	@mkdir -p $(BIN_DIR)
	GOOS=windows GOARCH=amd64 go build -o $(BIN_DIR)/site-tool-windows-amd64.exe .
	GOOS=windows GOARCH=arm64 go build -o $(BIN_DIR)/site-tool-windows-arm64.exe .
	GOOS=linux   GOARCH=amd64 go build -o $(BIN_DIR)/site-tool-linux-amd64 .
	GOOS=linux   GOARCH=arm64 go build -o $(BIN_DIR)/site-tool-linux-arm64 .
	GOOS=darwin  GOARCH=amd64 go build -o $(BIN_DIR)/site-tool-darwin-amd64 .
	GOOS=darwin  GOARCH=arm64 go build -o $(BIN_DIR)/site-tool-darwin-arm64 .

#
nuke:
	@echo "cleaning all site-tool builds"
	rm -rf $(BIN_DIR)
