.PHONY: build

tools:
	bash ./scripts/tools.sh

tests:
	bash ./scripts/tests.sh

tests-verbose:
	TEST_VERBOSE=true TEST_LOG_FORMAT=standard-verbose bash ./scripts/tests.sh

mod:
	bash ./scripts/mod.sh

lint:
	bash ./scripts/lint.sh

fix:
	bash ./scripts/fix.sh

generate-api:
	bash ./scripts/generate-api.sh

ci: mod lint tests

