NOW := $(shell echo "`date +%Y-%m-%d`")

#
# Display help
# 
define help_info
	@echo "\nUsage:\n"
	@echo "$$ make help            - Display this help text"
	@echo "$$ make buildall        - tinygo build all"
	@echo "$$ make clean           - cleanup stuf"
	@echo "\n"

endef

help:
	$(call help_info)

buildall:
	tinygo build -target=pico ./cmd/console/main.go
	tinygo build -target=pico ./cmd/de-driver/main.go
	tinygo build -target=pico ./cmd/ra-driver/main.go	
	tinygo build -target=pico ./cmd/handset/main.go

clean:
	

