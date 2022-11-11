NOW := $(shell echo "`date +%Y-%m-%d`")

#
# Display help
# 
define help_info
	@echo "\nUsage:\n"
	@echo "$$ make help            - Display this help text"
	@echo "$$ make build           - tinygo build all"
	@echo "$$ make clean           - cleanup stuf"
	@echo "\n"

endef

help:
	$(call help_info)

build: clean
	tinygo build -target=pico -o ./cmd/console/console.elf     ./cmd/console/main.go
	tinygo build -target=pico -o ./cmd/de-driver/de-driver.elf ./cmd/de-driver/main.go
	tinygo build -target=pico -o ./cmd/ra-driver/ra-driver.elf ./cmd/ra-driver/main.go	
	tinygo build -target=pico -o ./cmd/handset/handset.elf     ./cmd/handset/main.go

clean:
	find . -type f -name "*.elf" -exec rm {} \;


