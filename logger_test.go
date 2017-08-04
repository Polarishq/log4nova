package log4nova_test

import (
    "github.com/golang/mock/gomock"
    . "github.com/onsi/ginkgo"
    //. "github.com/onsi/gomega"
    "github.com/Polarishq/log4nova"
    "net/http/httptest"
    "github.com/sirupsen/logrus"
    "github.com/Polarishq/log4nova/mocks/logface-sdk-go/client/events"
    rtclient "github.com/go-openapi/runtime/client"
    "github.com/sirupsen/logrus/hooks/test"
    "github.com/Polarishq/logface-sdk-go/client/events"
    "github.com/Polarishq/logface-sdk-go/models"
)


var _ = Describe("Log4Nova Logger", func() {
    var (
        mockCtrl            *gomock.Controller
        mockEventsClient    *mock_events.MockClientInterface
        logger              *log4nova.NovaLogger
        testLogger          *logrus.Logger
        recorder            *httptest.ResponseRecorder
        clientID            string
        clientSecret        string
        host                string
    )

    BeforeEach(func() {
        mockCtrl = gomock.NewController(GinkgoT())
        mockEventsClient = mock_events.NewMockClientInterface(mockCtrl)
        testLogger, _ = test.NewNullLogger()
        clientID = "clientID"
        clientSecret = "clientSecret"
        host = "testhost"
        logger = log4nova.NewNovaLogger(mockEventsClient, testLogger, clientID, clientSecret, host)
        recorder = httptest.NewRecorder()
    })

    It("Should flush requests to the events endpoint", func() {
        //logger.Start()
        testData := "test data"
        logger.Write([]byte(testData))
        testEventParams := events.NewEventsParams()
        testEvent := models.Event{ Event: map[string]string{"raw": testData}}
        testEventParams.Events = models.Events{&testEvent}
        auth := rtclient.BasicAuth(clientID, clientSecret)

        ret := models.EventsReturn{Bytes: 1, Count: 1}
        eok := events.EventsOK{Payload: &ret}

        mockEventsClient.EXPECT().Events(testEventParams, auth).Return(&eok, nil).Times(1)
    })
})
