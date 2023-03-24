TESTS := $(wildcard tests/*.sh)

build:
	go build jlinker.go

test: build
	$(MAKE) $(TESTS)
	@printf '\e[32mPassed all tests \e[0m'

$(TESTS):
	@printf '\e[33mTesting %s\e[0m \n' $@
	@chmod a+x $@
	@./$@
	@printf '\e[32mSuccess\e[0m \n'

clean:
	go clean
	rm -rf out/

.PHONY: build clean test $(TESTS)