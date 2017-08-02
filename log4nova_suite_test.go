package log4nova

import (
    "testing"

    . "github.com/onsi/ginkgo"
    . "github.com/onsi/gomega"
)

func TestLog4Nova(t *testing.T) {
    RegisterFailHandler(Fail)
    RunSpecs(t, "Log4Nova Suite")
}
