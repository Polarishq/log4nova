package log4nova_test

import (
    "github.com/golang/mock/gomock"
    . "github.com/onsi/ginkgo"
    . "github.com/onsi/gomega"
    "github.com/splunknova/log4nova/mocks/log4nova"
    "github.com/splunknova/log4nova"
    "net/http"
    "net/http/httptest"
    "github.com/sirupsen/logrus"
)


var _ = Describe("Log4Nova Handler", func() {
    var (
        mockCtrl    *gomock.Controller
        mockLogger  *mock_log4nova.MockINovaLogger
        testHandler http.HandlerFunc
        handler     *log4nova.NovaHandler
        recorder      *httptest.ResponseRecorder
    )

    BeforeEach(func() {
        mockCtrl = gomock.NewController(GinkgoT())
        mockLogger = mock_log4nova.NewMockINovaLogger(mockCtrl)
        testHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            w.WriteHeader(200)
            w.Write([]byte("bar"))
        })
        handler = log4nova.NewNovaHandler(mockLogger, testHandler)
        recorder = httptest.NewRecorder()
    })

    It("Should log requests", func() {
        request, _ := http.NewRequest("GET", "http://example.com/foo", nil)
        logger := logrus.New()
        entry := logrus.NewEntry(logger)
        mockLogger.EXPECT().Start().Times(1)
        mockLogger.EXPECT().WithFields(gomock.Any()).Return(entry).Times(1)
        handler.ServeHTTP(recorder, request)
        Expect(recorder.Code).To(Equal(200))
    })
})
