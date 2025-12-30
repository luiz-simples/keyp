package service_test

import (
	"context"
	"encoding/binary"
	"errors"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.uber.org/mock/gomock"

	"github.com/luiz-simples/keyp.git/internal/service"
)

var _ = Describe("Handler Unit Tests", func() {
	var (
		ctrl          *gomock.Controller
		mockPersister *MockPersister
		handler       *service.Handler
		ctx           context.Context
	)

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
		mockPersister = NewMockPersister(ctrl)
		handler = service.NewHandler(mockPersister)
		ctx = context.Background()
	})

	AfterEach(func() {
		ctrl.Finish()
	})

	Describe("PING Command", func() {
		Context("when called without arguments", func() {
			It("should return PONG", func() {
				args := [][]byte{[]byte("PING")}
				results := handler.Apply(ctx, args)

				Expect(results).To(HaveLen(1))
				Expect(results[0].Error).To(BeNil())
				Expect(results[0].Response).To(Equal([]byte("PONG")))
			})
		})

		Context("when called with message", func() {
			It("should return the same message", func() {
				message := []byte("hello")
				args := [][]byte{[]byte("PING"), message}
				results := handler.Apply(ctx, args)

				Expect(results).To(HaveLen(1))
				Expect(results[0].Error).To(BeNil())
				Expect(results[0].Response).To(Equal(message))
			})
		})
	})

	Describe("SET Command", func() {
		Context("when storage succeeds", func() {
			It("should return OK", func() {
				key := []byte("testkey")
				value := []byte("testvalue")
				args := [][]byte{[]byte("SET"), key, value}

				mockPersister.EXPECT().
					Set(gomock.Any(), key, value).
					Return(nil)

				results := handler.Apply(ctx, args)

				Expect(results).To(HaveLen(1))
				Expect(results[0].Error).To(BeNil())
				Expect(results[0].Response).To(Equal([]byte("OK")))
			})
		})

		Context("when storage fails", func() {
			It("should return error", func() {
				key := []byte("testkey")
				value := []byte("testvalue")
				args := [][]byte{[]byte("SET"), key, value}
				expectedError := errors.New("storage error")

				mockPersister.EXPECT().
					Set(gomock.Any(), key, value).
					Return(expectedError)

				results := handler.Apply(ctx, args)

				Expect(results).To(HaveLen(1))
				Expect(results[0].Error).To(Equal(expectedError))
			})
		})

		Context("when context is canceled", func() {
			It("should return canceled error", func() {
				key := []byte("testkey")
				value := []byte("testvalue")
				args := [][]byte{[]byte("SET"), key, value}

				mockPersister.EXPECT().
					Set(gomock.Any(), key, value).
					Return(context.Canceled)

				results := handler.Apply(ctx, args)

				Expect(results).To(HaveLen(1))
				Expect(results[0].Error.Error()).To(ContainSubstring("operation canceled"))
			})
		})

		Context("with invalid arguments", func() {
			It("should return error for missing key", func() {
				args := [][]byte{[]byte("SET")}
				results := handler.Apply(ctx, args)

				Expect(results).To(HaveLen(1))
				Expect(results[0].Error).To(HaveOccurred())
				Expect(results[0].Error.Error()).To(ContainSubstring("wrong number of arguments"))
			})

			It("should return error for missing value", func() {
				args := [][]byte{[]byte("SET"), []byte("key")}
				results := handler.Apply(ctx, args)

				Expect(results).To(HaveLen(1))
				Expect(results[0].Error).To(HaveOccurred())
				Expect(results[0].Error.Error()).To(ContainSubstring("wrong number of arguments"))
			})
		})
	})

	Describe("GET Command", func() {
		Context("when key exists", func() {
			It("should return value", func() {
				key := []byte("testkey")
				expectedValue := []byte("testvalue")
				args := [][]byte{[]byte("GET"), key}

				mockPersister.EXPECT().
					Get(gomock.Any(), key).
					Return(expectedValue, nil)

				results := handler.Apply(ctx, args)

				Expect(results).To(HaveLen(1))
				Expect(results[0].Error).To(BeNil())
				Expect(results[0].Response).To(Equal(expectedValue))
			})
		})

		Context("when key does not exist", func() {
			It("should return nil response", func() {
				key := []byte("nonexistent")
				args := [][]byte{[]byte("GET"), key}

				mockPersister.EXPECT().
					Get(gomock.Any(), key).
					Return(nil, errors.New("key not found"))

				results := handler.Apply(ctx, args)

				Expect(results).To(HaveLen(1))
				Expect(results[0].Error).To(BeNil())
				Expect(results[0].Response).To(BeNil())
			})
		})

		Context("when storage fails", func() {
			It("should return error", func() {
				key := []byte("testkey")
				args := [][]byte{[]byte("GET"), key}
				expectedError := errors.New("storage error")

				mockPersister.EXPECT().
					Get(gomock.Any(), key).
					Return(nil, expectedError)

				results := handler.Apply(ctx, args)

				Expect(results).To(HaveLen(1))
				Expect(results[0].Error).To(Equal(expectedError))
			})
		})

		Context("when context is canceled", func() {
			It("should return canceled error", func() {
				key := []byte("testkey")
				args := [][]byte{[]byte("GET"), key}

				mockPersister.EXPECT().
					Get(gomock.Any(), key).
					Return(nil, context.Canceled)

				results := handler.Apply(ctx, args)

				Expect(results).To(HaveLen(1))
				Expect(results[0].Error.Error()).To(ContainSubstring("operation canceled"))
			})
		})
	})

	Describe("DEL Command", func() {
		Context("when deleting single key", func() {
			It("should return count of deleted keys", func() {
				key := []byte("testkey")
				args := [][]byte{[]byte("DEL"), key}
				expectedCount := uint32(1)

				mockPersister.EXPECT().
					Del(gomock.Any(), key).
					Return(expectedCount, nil)

				results := handler.Apply(ctx, args)

				Expect(results).To(HaveLen(1))
				Expect(results[0].Error).To(BeNil())

				actualCount := binary.LittleEndian.Uint32(results[0].Response)
				Expect(actualCount).To(Equal(expectedCount))
			})
		})

		Context("when deleting multiple keys", func() {
			It("should return count of deleted keys", func() {
				key1 := []byte("key1")
				key2 := []byte("key2")
				args := [][]byte{[]byte("DEL"), key1, key2}
				expectedCount := uint32(2)

				mockPersister.EXPECT().
					Del(gomock.Any(), key1, key2).
					Return(expectedCount, nil)

				results := handler.Apply(ctx, args)

				Expect(results).To(HaveLen(1))
				Expect(results[0].Error).To(BeNil())

				actualCount := binary.LittleEndian.Uint32(results[0].Response)
				Expect(actualCount).To(Equal(expectedCount))
			})
		})

		Context("when context is canceled", func() {
			It("should return canceled error", func() {
				key := []byte("testkey")
				args := [][]byte{[]byte("DEL"), key}

				mockPersister.EXPECT().
					Del(gomock.Any(), key).
					Return(uint32(0), context.Canceled)

				results := handler.Apply(ctx, args)

				Expect(results).To(HaveLen(1))
				Expect(results[0].Error.Error()).To(ContainSubstring("operation canceled"))
			})
		})
	})

	Describe("Unknown Command", func() {
		It("should return unknown command error", func() {
			args := [][]byte{[]byte("UNKNOWN")}
			results := handler.Apply(ctx, args)

			Expect(results).To(HaveLen(1))
			Expect(results[0].Error).To(HaveOccurred())
			Expect(results[0].Error.Error()).To(ContainSubstring("unknown command"))
		})
	})

	Describe("Empty Command", func() {
		It("should return empty command error", func() {
			args := [][]byte{}
			results := handler.Apply(ctx, args)

			Expect(results).To(HaveLen(1))
			Expect(results[0].Error).To(HaveOccurred())
			Expect(results[0].Error.Error()).To(ContainSubstring("empty command"))
		})
	})

	Describe("EXPIRE Command", func() {
		Context("when setting expiration", func() {
			It("should return OK", func() {
				key := []byte("testkey")
				seconds := []byte("60")
				args := [][]byte{[]byte("EXPIRE"), key, seconds}

				mockPersister.EXPECT().
					Expire(gomock.Any(), key, uint32(60)).
					Return()

				results := handler.Apply(ctx, args)

				Expect(results).To(HaveLen(1))
				Expect(results[0].Error).To(BeNil())
				Expect(results[0].Response).To(Equal([]byte("OK")))
			})
		})

		Context("when setting expiration with invalid seconds", func() {
			It("should handle gracefully", func() {
				key := []byte("testkey")
				seconds := []byte("invalid")
				args := [][]byte{[]byte("EXPIRE"), key, seconds}

				mockPersister.EXPECT().
					Expire(gomock.Any(), key, uint32(0)).
					Return()

				results := handler.Apply(ctx, args)

				Expect(results).To(HaveLen(1))
				Expect(results[0].Error).To(BeNil())
				Expect(results[0].Response).To(Equal([]byte("OK")))
			})
		})
	})

	Describe("PERSIST Command", func() {
		Context("when removing expiration", func() {
			It("should return OK", func() {
				key := []byte("testkey")
				args := [][]byte{[]byte("PERSIST"), key}

				mockPersister.EXPECT().
					Persist(gomock.Any(), key).
					Return()

				results := handler.Apply(ctx, args)

				Expect(results).To(HaveLen(1))
				Expect(results[0].Error).To(BeNil())
				Expect(results[0].Response).To(Equal([]byte("OK")))
			})
		})
	})

	Describe("TTL Command", func() {
		Context("when getting TTL", func() {
			It("should return TTL value", func() {
				key := []byte("testkey")
				args := [][]byte{[]byte("TTL"), key}
				expectedTTL := uint32(120)

				mockPersister.EXPECT().
					TTL(gomock.Any(), key).
					Return(expectedTTL)

				results := handler.Apply(ctx, args)

				Expect(results).To(HaveLen(1))
				Expect(results[0].Error).To(BeNil())

				expectedResponse := make([]byte, 4)
				binary.LittleEndian.PutUint32(expectedResponse, expectedTTL)
				Expect(results[0].Response).To(Equal(expectedResponse))
			})
		})

		Context("when key has no expiration", func() {
			It("should return -1", func() {
				key := []byte("testkey")
				args := [][]byte{[]byte("TTL"), key}
				expectedTTL := uint32(0xFFFFFFFF)

				mockPersister.EXPECT().
					TTL(gomock.Any(), key).
					Return(expectedTTL)

				results := handler.Apply(ctx, args)

				Expect(results).To(HaveLen(1))
				Expect(results[0].Error).To(BeNil())

				expectedResponse := make([]byte, 4)
				binary.LittleEndian.PutUint32(expectedResponse, expectedTTL)
				Expect(results[0].Response).To(Equal(expectedResponse))
			})
		})
	})

	Describe("SEL Command", func() {
		Context("when selecting database", func() {
			It("should return OK", func() {
				dbNumber := []byte("1")
				args := [][]byte{[]byte("SEL"), dbNumber}

				results := handler.Apply(ctx, args)

				Expect(results).To(HaveLen(1))
				Expect(results[0].Error).To(BeNil())
				Expect(results[0].Response).To(Equal([]byte("OK")))
			})
		})

		Context("when selecting database with invalid number", func() {
			It("should handle gracefully", func() {
				dbNumber := []byte("invalid")
				args := [][]byte{[]byte("SEL"), dbNumber}

				results := handler.Apply(ctx, args)

				Expect(results).To(HaveLen(1))
				Expect(results[0].Error).To(BeNil())
				Expect(results[0].Response).To(Equal([]byte("OK")))
			})
		})
	})

	Describe("MULTI Command", func() {
		Context("when starting transaction", func() {
			It("should enable multi mode", func() {
				args := [][]byte{[]byte("MULTI")}

				results := handler.Apply(ctx, args)

				Expect(results).To(HaveLen(1))
				Expect(results[0].Error).To(BeNil())
				Expect(results[0].Response).To(Equal([]byte("OK")))
			})
		})
	})

	Describe("EXEC Command", func() {
		Context("when executing queued commands", func() {
			It("should execute all queued commands", func() {
				key := []byte("testkey")
				value := []byte("testvalue")

				multiArgs := [][]byte{[]byte("MULTI")}
				handler.Apply(ctx, multiArgs)

				setArgs := [][]byte{[]byte("SET"), key, value}
				mockPersister.EXPECT().
					Set(gomock.Any(), key, value).
					Return(nil)

				handler.Apply(ctx, setArgs)

				execArgs := [][]byte{[]byte("EXEC")}
				results := handler.Apply(ctx, execArgs)

				Expect(results).To(HaveLen(1))
				Expect(results[0].Error).To(BeNil())
				Expect(results[0].Response).To(Equal([]byte("OK")))
			})
		})

		Context("when no transaction is active", func() {
			It("should return empty results", func() {
				args := [][]byte{[]byte("EXEC")}

				results := handler.Apply(ctx, args)

				Expect(results).To(HaveLen(0))
			})
		})
	})

	Describe("DISCARD Command", func() {
		Context("when discarding transaction", func() {
			It("should clear queued commands", func() {
				multiArgs := [][]byte{[]byte("MULTI")}
				handler.Apply(ctx, multiArgs)

				setArgs := [][]byte{[]byte("SET"), []byte("key"), []byte("value")}
				handler.Apply(ctx, setArgs)

				discardArgs := [][]byte{[]byte("DISCARD")}
				results := handler.Apply(ctx, discardArgs)

				Expect(results).To(HaveLen(1))
				Expect(results[0].Error).To(BeNil())
				Expect(results[0].Response).To(Equal([]byte("OK")))
			})
		})
	})

	Describe("Pool Management", func() {
		var (
			pool *service.Pool
		)

		BeforeEach(func() {
			pool = service.NewPool(mockPersister)
		})

		Describe("NewPool", func() {
			It("should create a new pool", func() {
				newPool := service.NewPool(mockPersister)
				Expect(newPool).NotTo(BeNil())
			})
		})

		Describe("Get", func() {
			It("should return a handler from pool", func() {
				handler := pool.Get(ctx)
				Expect(handler).NotTo(BeNil())
			})

			It("should return different handlers on multiple calls", func() {
				handler1 := pool.Get(ctx)
				handler2 := pool.Get(ctx)

				Expect(handler1).NotTo(BeNil())
				Expect(handler2).NotTo(BeNil())
			})
		})

		Describe("Free", func() {
			It("should return handler to pool and clear state", func() {
				handler := pool.Get(ctx)

				multiArgs := [][]byte{[]byte("MULTI")}
				handler.Apply(ctx, multiArgs)

				pool.Free(handler)

				newHandler := pool.Get(ctx)
				Expect(newHandler).NotTo(BeNil())
			})
		})
	})

	Describe("Handler Clear", func() {
		It("should reset handler state", func() {
			multiArgs := [][]byte{[]byte("MULTI")}
			handler.Apply(ctx, multiArgs)

			setArgs := [][]byte{[]byte("SET"), []byte("key"), []byte("value")}
			handler.Apply(ctx, setArgs)

			handler.Clear()

			execArgs := [][]byte{[]byte("EXEC")}
			results := handler.Apply(ctx, execArgs)

			Expect(results).To(HaveLen(0))
		})

		It("should clear multi state", func() {
			multiArgs := [][]byte{[]byte("MULTI")}
			handler.Apply(ctx, multiArgs)

			handler.Clear()

			execArgs := [][]byte{[]byte("EXEC")}
			results := handler.Apply(ctx, execArgs)

			Expect(results).To(HaveLen(0))
		})
	})
})
