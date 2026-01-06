package storage_test

import (
	"context"
	"os"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/luiz-simples/keyp.git/internal/domain"
	"github.com/luiz-simples/keyp.git/internal/storage"
)

var _ = Describe("Basic Storage Commands", func() {
	var (
		client  *storage.Client
		ctx     context.Context
		tempDir string
	)

	BeforeEach(func() {
		var err error
		tempDir, err = os.MkdirTemp("", "keyp-test-*")
		Expect(err).NotTo(HaveOccurred())

		client, err = storage.NewClient(tempDir)
		Expect(err).NotTo(HaveOccurred())

		ctx = context.WithValue(context.Background(), domain.DB, 0)
	})

	AfterEach(func() {
		if client != nil {
			client.Close()
		}
		os.RemoveAll(tempDir)
	})

	Describe("Exists", func() {
		It("should return false for non-existent key", func() {
			exists := client.Exists(ctx, []byte("nonexistent"))
			Expect(exists).To(BeFalse())
		})

		It("should return true for existing key", func() {
			// Set a key first
			err := client.Set(ctx, []byte("testkey"), []byte("testvalue"))
			Expect(err).NotTo(HaveOccurred())

			exists := client.Exists(ctx, []byte("testkey"))
			Expect(exists).To(BeTrue())
		})

		It("should return false for empty key", func() {
			exists := client.Exists(ctx, []byte(""))
			Expect(exists).To(BeFalse())
		})

		It("should return false for nil key", func() {
			exists := client.Exists(ctx, nil)
			Expect(exists).To(BeFalse())
		})
	})

	Describe("Incr", func() {
		It("should increment non-existent key to 1", func() {
			result, err := client.Incr(ctx, []byte("counter"))
			Expect(err).NotTo(HaveOccurred())
			Expect(result).To(Equal(int64(1)))
		})

		It("should increment existing integer key", func() {
			// Set initial value
			err := client.Set(ctx, []byte("counter"), []byte("5"))
			Expect(err).NotTo(HaveOccurred())

			result, err := client.Incr(ctx, []byte("counter"))
			Expect(err).NotTo(HaveOccurred())
			Expect(result).To(Equal(int64(6)))
		})

		It("should increment zero value", func() {
			err := client.Set(ctx, []byte("counter"), []byte("0"))
			Expect(err).NotTo(HaveOccurred())

			result, err := client.Incr(ctx, []byte("counter"))
			Expect(err).NotTo(HaveOccurred())
			Expect(result).To(Equal(int64(1)))
		})

		It("should increment negative value", func() {
			err := client.Set(ctx, []byte("counter"), []byte("-1"))
			Expect(err).NotTo(HaveOccurred())

			result, err := client.Incr(ctx, []byte("counter"))
			Expect(err).NotTo(HaveOccurred())
			Expect(result).To(Equal(int64(0)))
		})

		It("should return error for non-integer value", func() {
			err := client.Set(ctx, []byte("counter"), []byte("notanumber"))
			Expect(err).NotTo(HaveOccurred())

			_, err = client.Incr(ctx, []byte("counter"))
			Expect(err).To(HaveOccurred())
		})
	})

	Describe("IncrBy", func() {
		It("should increment non-existent key by specified amount", func() {
			result, err := client.IncrBy(ctx, []byte("counter"), 5)
			Expect(err).NotTo(HaveOccurred())
			Expect(result).To(Equal(int64(5)))
		})

		It("should increment existing key by specified amount", func() {
			err := client.Set(ctx, []byte("counter"), []byte("10"))
			Expect(err).NotTo(HaveOccurred())

			result, err := client.IncrBy(ctx, []byte("counter"), 3)
			Expect(err).NotTo(HaveOccurred())
			Expect(result).To(Equal(int64(13)))
		})

		It("should handle negative increment", func() {
			err := client.Set(ctx, []byte("counter"), []byte("10"))
			Expect(err).NotTo(HaveOccurred())

			result, err := client.IncrBy(ctx, []byte("counter"), -3)
			Expect(err).NotTo(HaveOccurred())
			Expect(result).To(Equal(int64(7)))
		})

		It("should increment by zero", func() {
			err := client.Set(ctx, []byte("counter"), []byte("5"))
			Expect(err).NotTo(HaveOccurred())

			result, err := client.IncrBy(ctx, []byte("counter"), 0)
			Expect(err).NotTo(HaveOccurred())
			Expect(result).To(Equal(int64(5)))
		})
	})

	Describe("Decr", func() {
		It("should decrement non-existent key to -1", func() {
			result, err := client.Decr(ctx, []byte("counter"))
			Expect(err).NotTo(HaveOccurred())
			Expect(result).To(Equal(int64(-1)))
		})

		It("should decrement existing integer key", func() {
			err := client.Set(ctx, []byte("counter"), []byte("5"))
			Expect(err).NotTo(HaveOccurred())

			result, err := client.Decr(ctx, []byte("counter"))
			Expect(err).NotTo(HaveOccurred())
			Expect(result).To(Equal(int64(4)))
		})

		It("should decrement zero value", func() {
			err := client.Set(ctx, []byte("counter"), []byte("0"))
			Expect(err).NotTo(HaveOccurred())

			result, err := client.Decr(ctx, []byte("counter"))
			Expect(err).NotTo(HaveOccurred())
			Expect(result).To(Equal(int64(-1)))
		})
	})

	Describe("DecrBy", func() {
		It("should decrement non-existent key by specified amount", func() {
			result, err := client.DecrBy(ctx, []byte("counter"), 5)
			Expect(err).NotTo(HaveOccurred())
			Expect(result).To(Equal(int64(-5)))
		})

		It("should decrement existing key by specified amount", func() {
			err := client.Set(ctx, []byte("counter"), []byte("10"))
			Expect(err).NotTo(HaveOccurred())

			result, err := client.DecrBy(ctx, []byte("counter"), 3)
			Expect(err).NotTo(HaveOccurred())
			Expect(result).To(Equal(int64(7)))
		})

		It("should handle negative decrement (acts like increment)", func() {
			err := client.Set(ctx, []byte("counter"), []byte("10"))
			Expect(err).NotTo(HaveOccurred())

			result, err := client.DecrBy(ctx, []byte("counter"), -3)
			Expect(err).NotTo(HaveOccurred())
			Expect(result).To(Equal(int64(13)))
		})
	})

	Describe("Append", func() {
		It("should append to non-existent key", func() {
			length := client.Append(ctx, []byte("key"), []byte("hello"))
			Expect(length).To(Equal(int64(5)))

			value, err := client.Get(ctx, []byte("key"))
			Expect(err).NotTo(HaveOccurred())
			Expect(value).To(Equal([]byte("hello")))
		})

		It("should append to existing key", func() {
			err := client.Set(ctx, []byte("key"), []byte("hello"))
			Expect(err).NotTo(HaveOccurred())

			length := client.Append(ctx, []byte("key"), []byte(" world"))
			Expect(length).To(Equal(int64(11)))

			value, err := client.Get(ctx, []byte("key"))
			Expect(err).NotTo(HaveOccurred())
			Expect(value).To(Equal([]byte("hello world")))
		})

		It("should handle empty append", func() {
			err := client.Set(ctx, []byte("key"), []byte("hello"))
			Expect(err).NotTo(HaveOccurred())

			length := client.Append(ctx, []byte("key"), []byte(""))
			Expect(length).To(Equal(int64(5)))

			value, err := client.Get(ctx, []byte("key"))
			Expect(err).NotTo(HaveOccurred())
			Expect(value).To(Equal([]byte("hello")))
		})

		It("should handle empty key", func() {
			length := client.Append(ctx, []byte(""), []byte("value"))
			Expect(length).To(Equal(int64(0)))
		})
	})

	Describe("FlushAll", func() {
		It("should clear all keys in database", func() {
			// Set some keys
			err := client.Set(ctx, []byte("key1"), []byte("value1"))
			Expect(err).NotTo(HaveOccurred())
			err = client.Set(ctx, []byte("key2"), []byte("value2"))
			Expect(err).NotTo(HaveOccurred())

			// Verify keys exist
			exists1 := client.Exists(ctx, []byte("key1"))
			exists2 := client.Exists(ctx, []byte("key2"))
			Expect(exists1).To(BeTrue())
			Expect(exists2).To(BeTrue())

			// Flush all
			err = client.FlushAll(ctx)
			Expect(err).NotTo(HaveOccurred())

			// Verify keys are gone
			exists1 = client.Exists(ctx, []byte("key1"))
			exists2 = client.Exists(ctx, []byte("key2"))
			Expect(exists1).To(BeFalse())
			Expect(exists2).To(BeFalse())
		})

		It("should handle empty database", func() {
			err := client.FlushAll(ctx)
			Expect(err).NotTo(HaveOccurred())
		})
	})
})
