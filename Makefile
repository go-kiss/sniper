default: rpc
	go build

clean:
	git clean -x -f -d

rpc:
	find rpc -name '*.proto' -exec protoc --twirp_out=. --go_out=. {} \;

.PHONY: test rpc
