.DEFAULT_GOAL := help
NOW := $(shell echo "`date +%Y-%m-%d`")

#
# Display help
# 
define help_info
	@echo "\nUsage:\n"
	@echo "$$ make help       - Display this help text"
	@echo "$$ make build      - tinygo build all"
	@echo "$$ make clean      - cleanup stuff"
	@echo "$$ make lint       - format all with golangci-lint"
	@echo "$$ make fmt        - format all with goimports"
	@echo "\n"

endef

.PHONY:help
help:
	$(call help_info)

.PHONY:clean
clean:
	find . -type f -name "*.elf" -exec rm {} \;

.PHONY:fmt
fmt:
	go fmt ./...
	goimports -l -w .

.PHONY:lint
lint: fmt
	# staticcheck -f stylish ./...
	golangci-lint run ./...

# vet: fmt
# 	go vet ./...
# .PHONY:vet

build: clean fmt
	tinygo build -target=pico -o ./cmd/console/console.elf     ./cmd/console/main.go
	tinygo build -target=pico -o ./cmd/de-driver/de-driver.elf ./cmd/de-driver/main.go
	tinygo build -target=pico -o ./cmd/ra-driver/ra-driver.elf ./cmd/ra-driver/main.go	
	tinygo build -target=pico -o ./cmd/handset/handset.elf     ./cmd/handset/main.go
.PHONY:build

#
# DEVTODO - add tests
# something like this:
# tinygo test -target=pico -run ^TestCell_SetChar$ github.com/tonygilkerson/astroeq/pkg/hid
