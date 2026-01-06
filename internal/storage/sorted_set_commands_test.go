package storage_test

import (
	"context"
	"os"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/luiz-simples/keyp.git/internal/domain"
	"github.com/luiz-simples/keyp.git/internal/storage"
)

var _ = Describe("Sorted Set Storage Commands", func() {
	var (
		client  *storage.Client
		ctx     context.Context
		tempDir string
	)

	BeforeEach(func() {
		var err error
		tempDir, err = os.MkdirTemp("", "keyp-test-zset-*")
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

	Describe("ZAdd", func() {
		It("should add single member to new sorted set", func() {
			count := client.ZAdd(ctx, []byte("zset"), 1.0, []byte("member1"))
			Expect(count).To(Equal(int64(1)))
		})

		It("should add member with different score", func() {
			count1 := client.ZAdd(ctx, []byte("zset"), 1.0, []byte("member1"))
			Expect(count1).To(Equal(int64(1)))

			count2 := client.ZAdd(ctx, []byte("zset"), 2.0, []byte("member2"))
			Expect(count2).To(Equal(int64(1)))
		})

		It("should update score of existing member", func() {
			// Add member with score 1.0
			count1 := client.ZAdd(ctx, []byte("zset"), 1.0, []byte("member1"))
			Expect(count1).To(Equal(int64(1)))

			// Update same member with score 2.0
			count2 := client.ZAdd(ctx, []byte("zset"), 2.0, []byte("member1"))
			Expect(count2).To(Equal(int64(0))) // No new member added, just updated
		})

		It("should handle zero score", func() {
			count := client.ZAdd(ctx, []byte("zset"), 0.0, []byte("member1"))
			Expect(count).To(Equal(int64(1)))
		})

		It("should handle negative score", func() {
			count := client.ZAdd(ctx, []byte("zset"), -1.5, []byte("member1"))
			Expect(count).To(Equal(int64(1)))
		})

		It("should handle empty key", func() {
			count := client.ZAdd(ctx, []byte(""), 1.0, []byte("member"))
			Expect(count).To(Equal(int64(0)))
		})

		It("should handle empty member", func() {
			count := client.ZAdd(ctx, []byte("zset"), 1.0, []byte(""))
			Expect(count).To(Equal(int64(0)))
		})
	})

	Describe("ZRange", func() {
		BeforeEach(func() {
			client.ZAdd(ctx, []byte("zset"), 1.0, []byte("member1"))
			client.ZAdd(ctx, []byte("zset"), 2.0, []byte("member2"))
			client.ZAdd(ctx, []byte("zset"), 3.0, []byte("member3"))
		})

		It("should return range of members by rank", func() {
			members, err := client.ZRange(ctx, []byte("zset"), 0, 1)
			Expect(err).NotTo(HaveOccurred())
			Expect(members).To(HaveLen(2))
			Expect(members[0]).To(Equal([]byte("member1"))) // Lowest score first
			Expect(members[1]).To(Equal([]byte("member2")))
		})

		It("should return all members with -1 end", func() {
			members, err := client.ZRange(ctx, []byte("zset"), 0, -1)
			Expect(err).NotTo(HaveOccurred())
			Expect(members).To(HaveLen(3))
			Expect(members[0]).To(Equal([]byte("member1")))
			Expect(members[1]).To(Equal([]byte("member2")))
			Expect(members[2]).To(Equal([]byte("member3")))
		})

		It("should handle negative indices", func() {
			members, err := client.ZRange(ctx, []byte("zset"), -2, -1)
			Expect(err).NotTo(HaveOccurred())
			Expect(members).To(HaveLen(2))
			Expect(members[0]).To(Equal([]byte("member2")))
			Expect(members[1]).To(Equal([]byte("member3")))
		})

		It("should return empty slice for non-existent sorted set", func() {
			members, err := client.ZRange(ctx, []byte("nonexistent"), 0, -1)
			Expect(err).NotTo(HaveOccurred())
			Expect(members).To(BeEmpty())
		})

		It("should handle empty key", func() {
			members, err := client.ZRange(ctx, []byte(""), 0, -1)
			Expect(err).NotTo(HaveOccurred())
			Expect(members).To(BeEmpty())
		})

		It("should handle out of bounds range", func() {
			members, err := client.ZRange(ctx, []byte("zset"), 10, 20)
			Expect(err).NotTo(HaveOccurred())
			Expect(members).To(BeEmpty())
		})
	})

	Describe("ZCount", func() {
		BeforeEach(func() {
			client.ZAdd(ctx, []byte("zset"), 1.0, []byte("member1"))
			client.ZAdd(ctx, []byte("zset"), 2.0, []byte("member2"))
			client.ZAdd(ctx, []byte("zset"), 3.0, []byte("member3"))
			client.ZAdd(ctx, []byte("zset"), 4.0, []byte("member4"))
		})

		It("should count members in score range", func() {
			count := client.ZCount(ctx, []byte("zset"), 1.0, 3.0)
			Expect(count).To(Equal(int64(3))) // member1, member2, member3
		})

		It("should count members with exact score match", func() {
			count := client.ZCount(ctx, []byte("zset"), 2.0, 2.0)
			Expect(count).To(Equal(int64(1))) // Only member2
		})

		It("should return 0 for range with no matches", func() {
			count := client.ZCount(ctx, []byte("zset"), 5.0, 6.0)
			Expect(count).To(Equal(int64(0)))
		})

		It("should handle negative scores", func() {
			client.ZAdd(ctx, []byte("zset"), -1.0, []byte("negative"))
			count := client.ZCount(ctx, []byte("zset"), -1.0, 1.0)
			Expect(count).To(Equal(int64(2))) // negative and member1
		})

		It("should return 0 for non-existent sorted set", func() {
			count := client.ZCount(ctx, []byte("nonexistent"), 1.0, 3.0)
			Expect(count).To(Equal(int64(0)))
		})

		It("should handle empty key", func() {
			count := client.ZCount(ctx, []byte(""), 1.0, 3.0)
			Expect(count).To(Equal(int64(0)))
		})

		It("should handle inverted range (min > max)", func() {
			count := client.ZCount(ctx, []byte("zset"), 3.0, 1.0)
			Expect(count).To(Equal(int64(0)))
		})
	})

	Describe("Sorted Set operations integration", func() {
		It("should maintain sorted order", func() {
			// Add members in random order
			client.ZAdd(ctx, []byte("zset"), 3.0, []byte("c"))
			client.ZAdd(ctx, []byte("zset"), 1.0, []byte("a"))
			client.ZAdd(ctx, []byte("zset"), 2.0, []byte("b"))

			// Should return in score order
			members, err := client.ZRange(ctx, []byte("zset"), 0, -1)
			Expect(err).NotTo(HaveOccurred())
			Expect(members).To(Equal([][]byte{
				[]byte("a"), // score 1.0
				[]byte("b"), // score 2.0
				[]byte("c"), // score 3.0
			}))
		})

		It("should handle score updates correctly", func() {
			// Add members
			client.ZAdd(ctx, []byte("zset"), 1.0, []byte("a"))
			client.ZAdd(ctx, []byte("zset"), 2.0, []byte("b"))

			// Update score of 'a' to be higher than 'b'
			client.ZAdd(ctx, []byte("zset"), 3.0, []byte("a"))

			// Should return in new order
			members, err := client.ZRange(ctx, []byte("zset"), 0, -1)
			Expect(err).NotTo(HaveOccurred())
			Expect(members).To(Equal([][]byte{
				[]byte("b"), // score 2.0
				[]byte("a"), // score 3.0 (updated)
			}))
		})

		It("should handle duplicate scores", func() {
			// Add members with same score
			client.ZAdd(ctx, []byte("zset"), 1.0, []byte("a"))
			client.ZAdd(ctx, []byte("zset"), 1.0, []byte("b"))
			client.ZAdd(ctx, []byte("zset"), 1.0, []byte("c"))

			// Should return all members
			members, err := client.ZRange(ctx, []byte("zset"), 0, -1)
			Expect(err).NotTo(HaveOccurred())
			Expect(members).To(HaveLen(3))

			// Count should work correctly
			count := client.ZCount(ctx, []byte("zset"), 1.0, 1.0)
			Expect(count).To(Equal(int64(3)))
		})
	})
})
