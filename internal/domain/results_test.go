package domain_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/luiz-simples/keyp.git/internal/domain"
)

var _ = Describe("Results", func() {
	Describe("NewResult", func() {
		It("should create a new cleared result", func() {
			result := domain.NewResult()

			Expect(result).NotTo(BeNil())
			Expect(result.Error).To(BeNil())
			Expect(result.Response).To(BeNil())
		})
	})

	Describe("Result methods", func() {
		var result *domain.Result

		BeforeEach(func() {
			result = domain.NewResult()
		})

		Describe("SetCanceled", func() {
			It("should set error to ErrCanceled and clear response", func() {
				result.Response = []byte("some data")

				returnedResult := result.SetCanceled()

				Expect(returnedResult).To(Equal(result)) // Should return self for chaining
				Expect(result.Error).To(Equal(domain.ErrCanceled))
				Expect(result.Response).To(BeNil())
			})
		})

		Describe("SetEmpty", func() {
			It("should set error to ErrEmpty and clear response", func() {
				result.Response = []byte("some data")

				returnedResult := result.SetEmpty()

				Expect(returnedResult).To(Equal(result)) // Should return self for chaining
				Expect(result.Error).To(Equal(domain.ErrEmpty))
				Expect(result.Response).To(BeNil())
			})
		})

		Describe("SetNil", func() {
			It("should clear both error and response", func() {
				result.Error = domain.ErrCanceled
				result.Response = []byte("some data")

				returnedResult := result.SetNil()

				Expect(returnedResult).To(Equal(result)) // Should return self for chaining
				Expect(result.Error).To(BeNil())
				Expect(result.Response).To(BeNil())
			})
		})

		Describe("Clear", func() {
			It("should clear both error and response", func() {
				result.Error = domain.ErrCanceled
				result.Response = []byte("some data")

				returnedResult := result.Clear()

				Expect(returnedResult).To(Equal(result)) // Should return self for chaining
				Expect(result.Error).To(BeNil())
				Expect(result.Response).To(BeNil())
			})
		})

		Describe("SetOK", func() {
			It("should set response to OK and clear error", func() {
				result.Error = domain.ErrCanceled

				returnedResult := result.SetOK()

				Expect(returnedResult).To(Equal(result)) // Should return self for chaining
				Expect(result.Error).To(BeNil())
				Expect(result.Response).To(Equal(domain.OK))
			})
		})
	})

	Describe("Method chaining", func() {
		It("should allow method chaining", func() {
			result := domain.NewResult().SetOK().Clear().SetCanceled()

			Expect(result.Error).To(Equal(domain.ErrCanceled))
			Expect(result.Response).To(BeNil())
		})
	})
})
