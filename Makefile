SDK_GO_MOCKS = mocks/logface-sdk-go/client/events
mock:
	go get -u github.com/golang/mock/mockgen
	mkdir -p $(SDK_GO_MOCKS)
	mockgen -source=vendor/github.com/Polarishq/logface-sdk-go/client/events/events_iface.go > $(SDK_GO_MOCKS)/mock.go
	mkdir -p mocks/log4nova
	mockgen github.com/Polarishq/log4nova INovaLogger > mocks/log4nova/mock_nova_logger.go

dependencies:
	@echo "Updating dependencies"
	govendor fetch +missing
	govendor add +external
	govendor sync
	govendor remove +unused
	@# For some reason these packages don't get pulled in automatically
	govendor fetch golang.org/x/text/unicode/norm
	govendor fetch golang.org/x/text/width
	govendor fetch golang.org/x/text/secure/bidirule

unittest:
	ginkgo -a -r -v $(PARALLEL_UNITTEST) -randomizeAllSpecs -randomizeSuites -progress -trace -cover