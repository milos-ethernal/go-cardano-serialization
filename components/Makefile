test-user:
	$(MAKE) -C user

test-batcher:
	$(MAKE) -C batcher

test-components:
	cat batcher/.env > .env \
    && echo >> .env \
    && cat relayer/.env >> .env
	go test -timeout 30s -v -run Test