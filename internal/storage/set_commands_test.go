package storage_test

import (
	"context"
	"os"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/luiz-simples/keyp.git/internal/domain"
	"github.com/luiz-simples/keyp.git/internal/storage"
)

var _ = Describe("Set Storage Commands", func() {
	var (
		client  *storage.Client
		ctx     context.Context
		tempDir string
	)

	BeforeEach(func() {
		var err error
		tempDir, err = os.MkdirTemp("", "keyp-test-set-*")
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

	Describe("SAdd", func() {
		It("should add single member to new set", func() {
			count := client.SAdd(ctx, []byte("set"), []byte("member1"))
			Expect(count).To(Equal(int64(1)))
		})

		It("should add multiple members to new set", func() {
			count := client.SAdd(ctx, []byte("set"), []byte("member1"), []byte("member2"), []byte("member3"))
			Expect(count).To(Equal(int64(3)))
		})

		It("should not add duplicate members", func() {
			// Add initial members
			count1 := client.SAdd(ctx, []byte("set"), []byte("member1"), []byte("member2"))
			Expect(count1).To(Equal(int64(2)))

			// Try to add duplicate and new member
			count2 := client.SAdd(ctx, []byte("set"), []byte("member1"), []byte("member3"))
			Expect(count2).To(Equal(int64(1))) // Only member3 should be added
		})

		It("should handle empty key", func() {
			count := client.SAdd(ctx, []byte(""), []byte("member"))
			Expect(count).To(Equal(int64(0)))
		})

		It("should handle empty members", func() {
			count := client.SAdd(ctx, []byte("set"))
			Expect(count).To(Equal(int64(0)))
		})
	})

	Describe("SRem", func() {
		BeforeEach(func() {
			client.SAdd(ctx, []byte("set"), []byte("member1"), []byte("member2"), []byte("member3"))
		})

		It("should remove existing member", func() {
			count := client.SRem(ctx, []byte("set"), []byte("member1"))
			Expect(count).To(Equal(int64(1)))
		})

		It("should remove multiple existing members", func() {
			count := client.SRem(ctx, []byte("set"), []byte("member1"), []byte("member2"))
			Expect(count).To(Equal(int64(2)))
		})

		It("should not remove non-existent member", func() {
			count := client.SRem(ctx, []byte("set"), []byte("nonexistent"))
			Expect(count).To(Equal(int64(0)))
		})

		It("should handle mix of existing and non-existent members", func() {
			count := client.SRem(ctx, []byte("set"), []byte("member1"), []byte("nonexistent"), []byte("member2"))
			Expect(count).To(Equal(int64(2))) // Only member1 and member2 should be removed
		})

		It("should handle non-existent set", func() {
			count := client.SRem(ctx, []byte("nonexistent"), []byte("member"))
			Expect(count).To(Equal(int64(0)))
		})

		It("should handle empty key", func() {
			count := client.SRem(ctx, []byte(""), []byte("member"))
			Expect(count).To(Equal(int64(0)))
		})

		It("should handle empty members", func() {
			count := client.SRem(ctx, []byte("set"))
			Expect(count).To(Equal(int64(0)))
		})
	})

	Describe("SIsMember", func() {
		BeforeEach(func() {
			client.SAdd(ctx, []byte("set"), []byte("member1"), []byte("member2"), []byte("member3"))
		})

		It("should return true for existing member", func() {
			isMember := client.SIsMember(ctx, []byte("set"), []byte("member1"))
			Expect(isMember).To(BeTrue())
		})

		It("should return false for non-existent member", func() {
			isMember := client.SIsMember(ctx, []byte("set"), []byte("nonexistent"))
			Expect(isMember).To(BeFalse())
		})

		It("should return false for non-existent set", func() {
			isMember := client.SIsMember(ctx, []byte("nonexistent"), []byte("member"))
			Expect(isMember).To(BeFalse())
		})

		It("should handle empty key", func() {
			isMember := client.SIsMember(ctx, []byte(""), []byte("member"))
			Expect(isMember).To(BeFalse())
		})

		It("should handle empty member", func() {
			isMember := client.SIsMember(ctx, []byte("set"), []byte(""))
			Expect(isMember).To(BeFalse())
		})
	})

	Describe("SMembers", func() {
		It("should return empty slice for non-existent set", func() {
			members, err := client.SMembers(ctx, []byte("nonexistent"))
			Expect(err).NotTo(HaveOccurred())
			Expect(members).To(BeEmpty())
		})

		It("should return all members of set", func() {
			client.SAdd(ctx, []byte("set"), []byte("member1"), []byte("member2"), []byte("member3"))

			members, err := client.SMembers(ctx, []byte("set"))
			Expect(err).NotTo(HaveOccurred())
			Expect(members).To(HaveLen(3))

			// Convert to strings for easier comparison (order may vary)
			memberStrings := make([]string, len(members))
			for i, member := range members {
				memberStrings[i] = string(member)
			}
			Expect(memberStrings).To(ContainElements("member1", "member2", "member3"))
		})

		It("should return empty slice for empty set", func() {
			// Create and then empty the set
			client.SAdd(ctx, []byte("set"), []byte("member1"))
			client.SRem(ctx, []byte("set"), []byte("member1"))

			members, err := client.SMembers(ctx, []byte("set"))
			Expect(err).NotTo(HaveOccurred())
			Expect(members).To(BeEmpty())
		})

		It("should handle empty key", func() {
			members, err := client.SMembers(ctx, []byte(""))
			Expect(err).NotTo(HaveOccurred())
			Expect(members).To(BeEmpty())
		})
	})

	Describe("Set operations integration", func() {
		It("should maintain set properties", func() {
			// Add members
			count := client.SAdd(ctx, []byte("set"), []byte("a"), []byte("b"), []byte("c"))
			Expect(count).To(Equal(int64(3)))

			// Check membership
			Expect(client.SIsMember(ctx, []byte("set"), []byte("a"))).To(BeTrue())
			Expect(client.SIsMember(ctx, []byte("set"), []byte("d"))).To(BeFalse())

			// Get all members
			members, err := client.SMembers(ctx, []byte("set"))
			Expect(err).NotTo(HaveOccurred())
			Expect(members).To(HaveLen(3))

			// Remove a member
			count = client.SRem(ctx, []byte("set"), []byte("b"))
			Expect(count).To(Equal(int64(1)))

			// Verify removal
			Expect(client.SIsMember(ctx, []byte("set"), []byte("b"))).To(BeFalse())

			members, err = client.SMembers(ctx, []byte("set"))
			Expect(err).NotTo(HaveOccurred())
			Expect(members).To(HaveLen(2))
		})
	})
})
