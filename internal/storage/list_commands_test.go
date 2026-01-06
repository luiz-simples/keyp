package storage_test

import (
	"context"
	"os"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/luiz-simples/keyp.git/internal/domain"
	"github.com/luiz-simples/keyp.git/internal/storage"
)

var _ = Describe("List Storage Commands", func() {
	var (
		client  *storage.Client
		ctx     context.Context
		tempDir string
	)

	BeforeEach(func() {
		var err error
		tempDir, err = os.MkdirTemp("", "keyp-test-list-*")
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

	Describe("LPush", func() {
		It("should push single element to new list", func() {
			length := client.LPush(ctx, []byte("list"), []byte("value1"))
			Expect(length).To(Equal(int64(1)))
		})

		It("should push multiple elements to new list", func() {
			length := client.LPush(ctx, []byte("list"), []byte("value1"), []byte("value2"), []byte("value3"))
			Expect(length).To(Equal(int64(3)))
		})

		It("should push elements to existing list", func() {
			// First push
			length1 := client.LPush(ctx, []byte("list"), []byte("value1"))
			Expect(length1).To(Equal(int64(1)))

			// Second push
			length2 := client.LPush(ctx, []byte("list"), []byte("value2"))
			Expect(length2).To(Equal(int64(2)))
		})

		It("should handle empty key", func() {
			length := client.LPush(ctx, []byte(""), []byte("value"))
			Expect(length).To(Equal(int64(0)))
		})

		It("should handle empty values", func() {
			length := client.LPush(ctx, []byte("list"))
			Expect(length).To(Equal(int64(0)))
		})
	})

	Describe("RPush", func() {
		It("should push single element to new list", func() {
			length := client.RPush(ctx, []byte("list"), []byte("value1"))
			Expect(length).To(Equal(int64(1)))
		})

		It("should push multiple elements to new list", func() {
			length := client.RPush(ctx, []byte("list"), []byte("value1"), []byte("value2"), []byte("value3"))
			Expect(length).To(Equal(int64(3)))
		})

		It("should push elements to existing list", func() {
			// First push
			length1 := client.RPush(ctx, []byte("list"), []byte("value1"))
			Expect(length1).To(Equal(int64(1)))

			// Second push
			length2 := client.RPush(ctx, []byte("list"), []byte("value2"))
			Expect(length2).To(Equal(int64(2)))
		})

		It("should handle empty key", func() {
			length := client.RPush(ctx, []byte(""), []byte("value"))
			Expect(length).To(Equal(int64(0)))
		})

		It("should handle empty values", func() {
			length := client.RPush(ctx, []byte("list"))
			Expect(length).To(Equal(int64(0)))
		})
	})

	Describe("LLen", func() {
		It("should return 0 for non-existent list", func() {
			length := client.LLen(ctx, []byte("nonexistent"))
			Expect(length).To(Equal(int64(0)))
		})

		It("should return correct length for existing list", func() {
			client.LPush(ctx, []byte("list"), []byte("value1"), []byte("value2"), []byte("value3"))

			length := client.LLen(ctx, []byte("list"))
			Expect(length).To(Equal(int64(3)))
		})

		It("should handle empty key", func() {
			length := client.LLen(ctx, []byte(""))
			Expect(length).To(Equal(int64(0)))
		})
	})

	Describe("LPop", func() {
		It("should return error for non-existent list", func() {
			_, err := client.LPop(ctx, []byte("nonexistent"))
			Expect(err).To(HaveOccurred())
		})

		It("should pop element from list", func() {
			client.LPush(ctx, []byte("list"), []byte("value1"), []byte("value2"))

			value, err := client.LPop(ctx, []byte("list"))
			Expect(err).NotTo(HaveOccurred())
			Expect(value).To(Equal([]byte("value2"))) // Last pushed should be first popped
		})

		It("should handle empty key", func() {
			_, err := client.LPop(ctx, []byte(""))
			Expect(err).To(HaveOccurred())
		})
	})

	Describe("RPop", func() {
		It("should return error for non-existent list", func() {
			_, err := client.RPop(ctx, []byte("nonexistent"))
			Expect(err).To(HaveOccurred())
		})

		It("should pop element from end of list", func() {
			client.LPush(ctx, []byte("list"), []byte("value1"), []byte("value2"))

			value, err := client.RPop(ctx, []byte("list"))
			Expect(err).NotTo(HaveOccurred())
			Expect(value).To(Equal([]byte("value1"))) // First pushed should be last popped
		})

		It("should handle empty key", func() {
			_, err := client.RPop(ctx, []byte(""))
			Expect(err).To(HaveOccurred())
		})
	})

	Describe("LIndex", func() {
		BeforeEach(func() {
			client.LPush(ctx, []byte("list"), []byte("value1"), []byte("value2"), []byte("value3"))
		})

		It("should return element at valid index", func() {
			value, err := client.LIndex(ctx, []byte("list"), 0)
			Expect(err).NotTo(HaveOccurred())
			Expect(value).To(Equal([]byte("value3"))) // Last pushed is at index 0
		})

		It("should return element at negative index", func() {
			value, err := client.LIndex(ctx, []byte("list"), -1)
			Expect(err).NotTo(HaveOccurred())
			Expect(value).To(Equal([]byte("value1"))) // Last element
		})

		It("should return error for out of bounds index", func() {
			_, err := client.LIndex(ctx, []byte("list"), 10)
			Expect(err).To(HaveOccurred())
		})

		It("should return error for non-existent list", func() {
			_, err := client.LIndex(ctx, []byte("nonexistent"), 0)
			Expect(err).To(HaveOccurred())
		})

		It("should handle empty key", func() {
			_, err := client.LIndex(ctx, []byte(""), 0)
			Expect(err).To(HaveOccurred())
		})
	})

	Describe("LSet", func() {
		BeforeEach(func() {
			client.LPush(ctx, []byte("list"), []byte("value1"), []byte("value2"), []byte("value3"))
		})

		It("should set element at valid index", func() {
			err := client.LSet(ctx, []byte("list"), 0, []byte("newvalue"))
			Expect(err).NotTo(HaveOccurred())

			value, err := client.LIndex(ctx, []byte("list"), 0)
			Expect(err).NotTo(HaveOccurred())
			Expect(value).To(Equal([]byte("newvalue")))
		})

		It("should set element at negative index", func() {
			err := client.LSet(ctx, []byte("list"), -1, []byte("newvalue"))
			Expect(err).NotTo(HaveOccurred())

			value, err := client.LIndex(ctx, []byte("list"), -1)
			Expect(err).NotTo(HaveOccurred())
			Expect(value).To(Equal([]byte("newvalue")))
		})

		It("should return error for out of bounds index", func() {
			err := client.LSet(ctx, []byte("list"), 10, []byte("newvalue"))
			Expect(err).To(HaveOccurred())
		})

		It("should return error for non-existent list", func() {
			err := client.LSet(ctx, []byte("nonexistent"), 0, []byte("newvalue"))
			Expect(err).To(HaveOccurred())
		})

		It("should handle empty key", func() {
			err := client.LSet(ctx, []byte(""), 0, []byte("newvalue"))
			Expect(err).To(HaveOccurred())
		})
	})

	Describe("LRange", func() {
		BeforeEach(func() {
			client.LPush(ctx, []byte("list"), []byte("value1"), []byte("value2"), []byte("value3"))
		})

		It("should return range of elements", func() {
			values, err := client.LRange(ctx, []byte("list"), 0, 1)
			Expect(err).NotTo(HaveOccurred())
			Expect(values).To(HaveLen(2))
			Expect(values[0]).To(Equal([]byte("value3")))
			Expect(values[1]).To(Equal([]byte("value2")))
		})

		It("should return all elements with -1 end", func() {
			values, err := client.LRange(ctx, []byte("list"), 0, -1)
			Expect(err).NotTo(HaveOccurred())
			Expect(values).To(HaveLen(3))
		})

		It("should handle negative indices", func() {
			values, err := client.LRange(ctx, []byte("list"), -2, -1)
			Expect(err).NotTo(HaveOccurred())
			Expect(values).To(HaveLen(2))
		})

		It("should return empty slice for non-existent list", func() {
			values, err := client.LRange(ctx, []byte("nonexistent"), 0, -1)
			Expect(err).NotTo(HaveOccurred())
			Expect(values).To(BeEmpty())
		})

		It("should handle empty key", func() {
			values, err := client.LRange(ctx, []byte(""), 0, -1)
			Expect(err).NotTo(HaveOccurred())
			Expect(values).To(BeEmpty())
		})

		It("should handle out of bounds range", func() {
			values, err := client.LRange(ctx, []byte("list"), 10, 20)
			Expect(err).NotTo(HaveOccurred())
			Expect(values).To(BeEmpty())
		})
	})
})
