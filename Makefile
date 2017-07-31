include skel/Makefile
BOUNCER_MOCKS = mocks/bouncer/client/events

mock:
	go get -u github.com/golang/mock/mockgen
	mkdir -p $(BOUNCER_MOCKS)
	mockgen -source=vendor/github.com/Polarishq/bouncer/client/events/events_iface.go > $(BOUNCER_MOCKS)/mock.go
	mkdir -p mocks/log4nova
	mockgen github.com/Polarishq/log4nova INovaLogger > mocks/log4nova/mock_nova_logger.go
