BOUNCER_MOCKS = mocks/bouncer/client/events
mock:
	go get -u github.com/golang/mock/mockgen
	mkdir -p $(BOUNCER_MOCKS)
	mockgen -source=vendor/github.com/Polarishq/bouncer/client/events/events_iface.go > $(BOUNCER_MOCKS)/mock.go
	mkdir -p mocks/log4nova
	mockgen github.com/Polarishq/log4nova INovaLogger > mocks/log4nova/mock_nova_logger.go

dependencies:
#
#__Fixes up govendor dependencies__
#
#* Adds missing packages
#* Removes unused packages
#
	@echo "Updating dependencies"
	govendor fetch +missing
	govendor add +external
	govendor sync
	govendor remove +unused
	@# For some reason these packages don't get pulled in automatically
	govendor fetch golang.org/x/text/unicode/norm
	govendor fetch golang.org/x/text/width
	govendor fetch golang.org/x/text/secure/bidirule