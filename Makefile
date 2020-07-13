RPC_PROTOS := $(shell find rpc -name '*.proto')
OTHER_PROTOS := $(shell find util -name '*.proto')

PBGENS := $(RPC_PROTOS:.proto=.pb.go) $(RPC_PROTOS:.proto=.twirp.go) \
	$(OTHER_PROTOS:.proto=.pb.go)

.PRECIOUS: $(PBGENS)

%.pb.go %.twirp.go: %.proto
	protoc --twirp_out=. --go_out=. $<

default: $(PBGENS)
	go build

rpc: $(PBGENS)

.PHONY: test rpc
