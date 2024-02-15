ifeq ($(OS), Windows_NT)
	DEL := del
else
	DEL := rm
endif

test-multisig-tx :
	cd tx && go test -timeout 30s -v -run TestMultisigTx

test-simple-tx :
	cd tx && go test -timeout 30s -v -run TestSimpleTx

test-user:
	$(MAKE) -C components test-user

test-batcher:
	$(MAKE) -C components test-batcher

test-components:
	$(MAKE) -C components test-components

test-e2e:
	$(MAKE) -C e2e test-e2e
	
all : test

test : coverage clean

coverage: 
	go test ./... -v -cover -coverprofile="coverage.out"
	go tool cover -html="coverage.out"

clean :
	$(DEL) coverage.out
	go clean
