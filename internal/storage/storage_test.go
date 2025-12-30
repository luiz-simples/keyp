package storage_test

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestStorage(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Storage Suite", Label("storage"))
}

func createUniqueTestDir(prefix string) string {
	timestamp := time.Now().UnixNano()
	pid := os.Getpid()
	return filepath.Join(os.TempDir(), fmt.Sprintf("keyp-storage-%s-%d-%d", prefix, pid, timestamp))
}

func cleanupTestDir(dir string) {
	if dir != "" {
		os.RemoveAll(dir)
	}
}
