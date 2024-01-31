ifeq ($(OS), Windows_NT)
	DEL := del
else
	DEL := rm
endif

test-multisig-tx :
	cd tx && go test -timeout 30s -v -run TestMultisigTx

test-simple-tx :
	cd tx && go test -timeout 30s -v -run TestSimpleTx
	
all : test

test : coverage clean

coverage: 
	go test ./... -v -cover -coverprofile="coverage.out"
	go tool cover -html="coverage.out"

clean :
	$(DEL) coverage.out
	go clean
