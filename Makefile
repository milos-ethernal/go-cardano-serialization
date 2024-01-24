ifeq ($(OS), Windows_NT)
	DEL := del
else
	DEL := rm
endif

test-tx :
	cd tx && go test -timeout 30s -v -run TestMultisigTx

all : test

test : coverage clean

coverage: 
	go test ./... -v -cover -coverprofile="coverage.out"
	go tool cover -html="coverage.out"

clean :
	$(DEL) coverage.out
	go clean
