package service_test

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestService(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Service Suite", Label("service"))
}

func createUniqueTestDir(prefix string) string {
	timestamp := time.Now().UnixNano()
	pid := os.Getpid()
	return filepath.Join(os.TempDir(), fmt.Sprintf("keyp-%s-%d-%d", prefix, pid, timestamp))
}

func cleanupTestDir(dir string) {
	if dir != "" {
		os.RemoveAll(dir)
	}
}
