TESTS := $(wildcard tests/*.sh)

build:
	@go build jlinker.go
	@ln -sf jlinker ld

test: build
	@CC="riscv64-linux-gnu-gcc" \
	$(MAKE) $(TESTS)
	@printf '\e[32mPassed all tests \e[0m'

$(TESTS):
	@printf '\e[33mTesting %s\e[0m \n' $@
	@chmod a+x $@
	@./$@
	@printf '\e[32mSuccess\e[0m \n'

clean:
	@go clean
	@rm -rf out/
	@rm -rf ld

.PHONY: build clean test $(TESTS)