package storage_test

import (
	"os"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/luiz-simples/keyp.git/internal/storage"
)

func TestStorage(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Storage Suite")
}

var _ = Describe("LMDB Storage", func() {
	var (
		lmdbStorage *storage.LMDBStorage
		tmpDir      string
	)

	BeforeEach(func() {
		// Set test mode to disable logging during tests
		os.Setenv("KEYP_TEST_MODE", "true")

		var err error
		tmpDir, err = os.MkdirTemp("", "lmdb-test-*")
		Expect(err).NotTo(HaveOccurred())

		lmdbStorage, err = storage.NewLMDBStorage(tmpDir)
		Expect(err).NotTo(HaveOccurred())
	})

	AfterEach(func() {
		if lmdbStorage != nil {
			lmdbStorage.Close()
		}
		if tmpDir != "" {
			os.RemoveAll(tmpDir)
		}
	})

	Describe("Set", func() {
		It("should store a key-value pair", func() {
			err := lmdbStorage.Set([]byte("key"), []byte("value"))
			Expect(err).NotTo(HaveOccurred())
		})

		It("should reject empty keys", func() {
			err := lmdbStorage.Set([]byte{}, []byte("value"))
			Expect(err).To(Equal(storage.ErrEmptyKey))
		})

		It("should reject oversized keys", func() {
			largeKey := make([]byte, storage.MaxKeySize+1)
			err := lmdbStorage.Set(largeKey, []byte("value"))
			Expect(err).To(Equal(storage.ErrKeyTooLarge))
		})

		It("should accept maximum size keys", func() {
			maxKey := make([]byte, storage.MaxKeySize)
			for i := range maxKey {
				maxKey[i] = byte('a')
			}
			err := lmdbStorage.Set(maxKey, []byte("value"))
			Expect(err).NotTo(HaveOccurred())
		})

		It("should overwrite existing keys", func() {
			key := []byte("key")
			err := lmdbStorage.Set(key, []byte("value1"))
			Expect(err).NotTo(HaveOccurred())

			err = lmdbStorage.Set(key, []byte("value2"))
			Expect(err).NotTo(HaveOccurred())

			value, err := lmdbStorage.Get(key)
			Expect(err).NotTo(HaveOccurred())
			Expect(value).To(Equal([]byte("value2")))
		})
	})

	Describe("Get", func() {
		It("should retrieve stored value", func() {
			key := []byte("test-key")
			value := []byte("test-value")

			err := lmdbStorage.Set(key, value)
			Expect(err).NotTo(HaveOccurred())

			retrieved, err := lmdbStorage.Get(key)
			Expect(err).NotTo(HaveOccurred())
			Expect(retrieved).To(Equal(value))
		})

		It("should return error for non-existent key", func() {
			_, err := lmdbStorage.Get([]byte("non-existent"))
			Expect(err).To(Equal(storage.ErrKeyNotFound))
		})

		It("should reject empty keys", func() {
			_, err := lmdbStorage.Get([]byte{})
			Expect(err).To(Equal(storage.ErrEmptyKey))
		})

		It("should reject oversized keys", func() {
			largeKey := make([]byte, storage.MaxKeySize+1)
			_, err := lmdbStorage.Get(largeKey)
			Expect(err).To(Equal(storage.ErrKeyTooLarge))
		})

		It("should handle binary data", func() {
			key := []byte("binary-key")
			value := []byte{0x00, 0xFF, 0x01, 0xFE, 0x02}

			err := lmdbStorage.Set(key, value)
			Expect(err).NotTo(HaveOccurred())

			retrieved, err := lmdbStorage.Get(key)
			Expect(err).NotTo(HaveOccurred())
			Expect(retrieved).To(Equal(value))
		})
	})

	Describe("Del", func() {
		It("should delete existing key", func() {
			key := []byte("key-to-delete")
			err := lmdbStorage.Set(key, []byte("value"))
			Expect(err).NotTo(HaveOccurred())

			deleted, err := lmdbStorage.Del(key)
			Expect(err).NotTo(HaveOccurred())
			Expect(deleted).To(Equal(1))

			_, err = lmdbStorage.Get(key)
			Expect(err).To(Equal(storage.ErrKeyNotFound))
		})

		It("should return 0 for non-existent key", func() {
			deleted, err := lmdbStorage.Del([]byte("non-existent"))
			Expect(err).NotTo(HaveOccurred())
			Expect(deleted).To(Equal(0))
		})

		It("should delete multiple keys", func() {
			keys := [][]byte{
				[]byte("key1"),
				[]byte("key2"),
				[]byte("key3"),
			}

			for _, key := range keys {
				err := lmdbStorage.Set(key, []byte("value"))
				Expect(err).NotTo(HaveOccurred())
			}

			deleted, err := lmdbStorage.Del(keys...)
			Expect(err).NotTo(HaveOccurred())
			Expect(deleted).To(Equal(3))

			for _, key := range keys {
				_, err := lmdbStorage.Get(key)
				Expect(err).To(Equal(storage.ErrKeyNotFound))
			}
		})

		It("should handle mixed existent and non-existent keys", func() {
			existingKey := []byte("existing")
			err := lmdbStorage.Set(existingKey, []byte("value"))
			Expect(err).NotTo(HaveOccurred())

			deleted, err := lmdbStorage.Del(
				existingKey,
				[]byte("non-existent1"),
				[]byte("non-existent2"),
			)
			Expect(err).NotTo(HaveOccurred())
			Expect(deleted).To(Equal(1))
		})

		It("should skip invalid keys", func() {
			validKey := []byte("valid")
			err := lmdbStorage.Set(validKey, []byte("value"))
			Expect(err).NotTo(HaveOccurred())

			largeKey := make([]byte, storage.MaxKeySize+1)

			deleted, err := lmdbStorage.Del(
				validKey,
				[]byte{},
				largeKey,
			)
			Expect(err).NotTo(HaveOccurred())
			Expect(deleted).To(Equal(1))
		})
	})
})
