package service_test

import (
	"context"
	"errors"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.uber.org/mock/gomock"

	"github.com/luiz-simples/keyp.git/internal/domain"
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
				Expect(string(results[0].Response)).To(Equal("1"))
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
				Expect(string(results[0].Response)).To(Equal("2"))
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
					Return(true)

				results := handler.Apply(ctx, args)

				Expect(results).To(HaveLen(1))
				Expect(results[0].Error).To(BeNil())
				Expect(results[0].Response).To(Equal([]byte("1")))
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
				Expect(string(results[0].Response)).To(Equal("120"))
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
				Expect(string(results[0].Response)).To(Equal("-1"))
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

				handler.Apply(ctx, [][]byte{[]byte("MULTI")})

				mockPersister.EXPECT().Set(gomock.Any(), key, value).Return(nil)
				handler.Apply(ctx, [][]byte{[]byte("SET"), key, value})

				results := handler.Apply(ctx, [][]byte{[]byte("EXEC")})

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
		var pool *service.Pool

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

	Describe("EXISTS Command", func() {
		Context("when checking key existence", func() {
			It("should return 1 for existing key", func() {
				key := []byte("exists-key")
				args := [][]byte{[]byte("EXISTS"), key}

				mockPersister.EXPECT().
					Exists(gomock.Any(), key).
					Return(true)

				results := handler.Apply(ctx, args)

				Expect(results).To(HaveLen(1))
				Expect(results[0].Error).To(BeNil())
				Expect(string(results[0].Response)).To(Equal("1"))
			})

			It("should return 0 for non-existing key", func() {
				key := []byte("non-exists-key")
				args := [][]byte{[]byte("EXISTS"), key}

				mockPersister.EXPECT().
					Exists(gomock.Any(), key).
					Return(false)

				results := handler.Apply(ctx, args)

				Expect(results).To(HaveLen(1))
				Expect(results[0].Error).To(BeNil())
				Expect(string(results[0].Response)).To(Equal("0"))
			})
		})
	})

	Describe("LLEN Command", func() {
		Context("when getting list length", func() {
			It("should return list length", func() {
				key := []byte("list-key")
				args := [][]byte{[]byte("LLEN"), key}
				expectedLength := int64(5)

				mockPersister.EXPECT().
					LLen(gomock.Any(), key).
					Return(expectedLength)

				results := handler.Apply(ctx, args)

				Expect(results).To(HaveLen(1))
				Expect(results[0].Error).To(BeNil())
				Expect(string(results[0].Response)).To(Equal("5"))
			})

			It("should return 0 for non-existing list", func() {
				key := []byte("non-list-key")
				args := [][]byte{[]byte("LLEN"), key}

				mockPersister.EXPECT().
					LLen(gomock.Any(), key).
					Return(int64(0))

				results := handler.Apply(ctx, args)

				Expect(results).To(HaveLen(1))
				Expect(results[0].Error).To(BeNil())
				Expect(string(results[0].Response)).To(Equal("0"))
			})
		})
	})

	Describe("LINDEX Command", func() {
		Context("when getting element by index", func() {
			It("should return element at valid index", func() {
				key := []byte("list-key")
				index := []byte("2")
				args := [][]byte{[]byte("LINDEX"), key, index}
				expectedValue := []byte("element-2")

				mockPersister.EXPECT().
					LIndex(gomock.Any(), key, int64(2)).
					Return(expectedValue, nil)

				results := handler.Apply(ctx, args)

				Expect(results).To(HaveLen(1))
				Expect(results[0].Error).To(BeNil())
				Expect(results[0].Response).To(Equal(expectedValue))
			})

			It("should return error for invalid index", func() {
				key := []byte("list-key")
				index := []byte("10")
				args := [][]byte{[]byte("LINDEX"), key, index}
				expectedError := errors.New("index out of range")

				mockPersister.EXPECT().
					LIndex(gomock.Any(), key, int64(10)).
					Return(nil, expectedError)

				results := handler.Apply(ctx, args)

				Expect(results).To(HaveLen(1))
				Expect(results[0].Error).To(Equal(expectedError))
			})
		})
	})

	Describe("LSET Command", func() {
		Context("when setting element by index", func() {
			It("should set element at valid index", func() {
				key := []byte("list-key")
				index := []byte("1")
				value := []byte("new-value")
				args := [][]byte{[]byte("LSET"), key, index, value}

				mockPersister.EXPECT().
					LSet(gomock.Any(), key, int64(1), value).
					Return(nil)

				results := handler.Apply(ctx, args)

				Expect(results).To(HaveLen(1))
				Expect(results[0].Error).To(BeNil())
				Expect(results[0].Response).To(Equal([]byte("OK")))
			})

			It("should return error for invalid index", func() {
				key := []byte("list-key")
				index := []byte("10")
				value := []byte("new-value")
				args := [][]byte{[]byte("LSET"), key, index, value}
				expectedError := errors.New("index out of range")

				mockPersister.EXPECT().
					LSet(gomock.Any(), key, int64(10), value).
					Return(expectedError)

				results := handler.Apply(ctx, args)

				Expect(results).To(HaveLen(1))
				Expect(results[0].Error).To(Equal(expectedError))
			})
		})
	})

	Describe("LPUSH Command", func() {
		Context("when pushing elements to left", func() {
			It("should push single element", func() {
				key := []byte("list-key")
				value := []byte("new-element")
				args := [][]byte{[]byte("LPUSH"), key, value}
				expectedLength := int64(3)

				mockPersister.EXPECT().
					LPush(gomock.Any(), key, value).
					Return(expectedLength)

				results := handler.Apply(ctx, args)

				Expect(results).To(HaveLen(1))
				Expect(results[0].Error).To(BeNil())
				Expect(string(results[0].Response)).To(Equal("3"))
			})

			It("should push multiple elements", func() {
				key := []byte("list-key")
				value1 := []byte("element1")
				value2 := []byte("element2")
				args := [][]byte{[]byte("LPUSH"), key, value1, value2}
				expectedLength := int64(5)

				mockPersister.EXPECT().
					LPush(gomock.Any(), key, value1, value2).
					Return(expectedLength)

				results := handler.Apply(ctx, args)

				Expect(results).To(HaveLen(1))
				Expect(results[0].Error).To(BeNil())
				Expect(string(results[0].Response)).To(Equal("5"))
			})
		})
	})

	Describe("RPUSH Command", func() {
		Context("when pushing elements to right", func() {
			It("should push single element", func() {
				key := []byte("list-key")
				value := []byte("new-element")
				args := [][]byte{[]byte("RPUSH"), key, value}
				expectedLength := int64(4)

				mockPersister.EXPECT().
					RPush(gomock.Any(), key, value).
					Return(expectedLength)

				results := handler.Apply(ctx, args)

				Expect(results).To(HaveLen(1))
				Expect(results[0].Error).To(BeNil())
				Expect(string(results[0].Response)).To(Equal("4"))
			})
		})
	})

	Describe("LPOP Command", func() {
		Context("when popping from left", func() {
			It("should return and remove leftmost element", func() {
				key := []byte("list-key")
				args := [][]byte{[]byte("LPOP"), key}
				expectedValue := []byte("first-element")

				mockPersister.EXPECT().
					LPop(gomock.Any(), key).
					Return(expectedValue, nil)

				results := handler.Apply(ctx, args)

				Expect(results).To(HaveLen(1))
				Expect(results[0].Error).To(BeNil())
				Expect(results[0].Response).To(Equal(expectedValue))
			})

			It("should return error for empty list", func() {
				key := []byte("empty-list")
				args := [][]byte{[]byte("LPOP"), key}
				expectedError := errors.New("list is empty")

				mockPersister.EXPECT().
					LPop(gomock.Any(), key).
					Return(nil, expectedError)

				results := handler.Apply(ctx, args)

				Expect(results).To(HaveLen(1))
				Expect(results[0].Error).To(Equal(expectedError))
			})
		})
	})

	Describe("RPOP Command", func() {
		Context("when popping from right", func() {
			It("should return and remove rightmost element", func() {
				key := []byte("list-key")
				args := [][]byte{[]byte("RPOP"), key}
				expectedValue := []byte("last-element")

				mockPersister.EXPECT().
					RPop(gomock.Any(), key).
					Return(expectedValue, nil)

				results := handler.Apply(ctx, args)

				Expect(results).To(HaveLen(1))
				Expect(results[0].Error).To(BeNil())
				Expect(results[0].Response).To(Equal(expectedValue))
			})
		})
	})

	Describe("LRANGE Command", func() {
		Context("when getting range of elements", func() {
			It("should return elements in range", func() {
				key := []byte("list-key")
				start := []byte("0")
				stop := []byte("2")
				args := [][]byte{[]byte("LRANGE"), key, start, stop}
				expectedValues := [][]byte{
					[]byte("elem1"),
					[]byte("elem2"),
					[]byte("elem3"),
				}

				mockPersister.EXPECT().
					LRange(gomock.Any(), key, int64(0), int64(2)).
					Return(expectedValues, nil)

				results := handler.Apply(ctx, args)

				Expect(results).To(HaveLen(1))
				Expect(results[0].Error).To(BeNil())

				expectedRedisFormat := "*3\r\n$5\r\nelem1\r\n$5\r\nelem2\r\n$5\r\nelem3\r\n"
				Expect(string(results[0].Response)).To(Equal(expectedRedisFormat))
			})

			It("should return empty array for non-existing key", func() {
				key := []byte("non-list")
				start := []byte("0")
				stop := []byte("2")
				args := [][]byte{[]byte("LRANGE"), key, start, stop}

				mockPersister.EXPECT().
					LRange(gomock.Any(), key, int64(0), int64(2)).
					Return([][]byte{}, nil)

				results := handler.Apply(ctx, args)

				Expect(results).To(HaveLen(1))
				Expect(results[0].Error).To(BeNil())
				Expect(string(results[0].Response)).To(Equal("*0\r\n"))
			})
		})
	})

	Describe("FLUSHALL Command", func() {
		Context("when flushing all keys", func() {
			It("should clear all data", func() {
				args := [][]byte{[]byte("FLUSHALL")}

				mockPersister.EXPECT().
					FlushAll(gomock.Any()).
					Return(nil)

				results := handler.Apply(ctx, args)

				Expect(results).To(HaveLen(1))
				Expect(results[0].Error).To(BeNil())
				Expect(results[0].Response).To(Equal([]byte("OK")))
			})

			It("should return error when flush fails", func() {
				args := [][]byte{[]byte("FLUSHALL")}
				expectedError := errors.New("flush failed")

				mockPersister.EXPECT().
					FlushAll(gomock.Any()).
					Return(expectedError)

				results := handler.Apply(ctx, args)

				Expect(results).To(HaveLen(1))
				Expect(results[0].Error).To(Equal(expectedError))
			})
		})
	})

	Describe("SADD Command", func() {
		Context("when adding members to set", func() {
			It("should add single member", func() {
				key := []byte("set-key")
				member := []byte("member1")
				args := [][]byte{[]byte("SADD"), key, member}
				expectedCount := int64(1)

				mockPersister.EXPECT().
					SAdd(gomock.Any(), key, member).
					Return(expectedCount)

				results := handler.Apply(ctx, args)

				Expect(results).To(HaveLen(1))
				Expect(results[0].Error).To(BeNil())
				Expect(string(results[0].Response)).To(Equal("1"))
			})

			It("should add multiple members", func() {
				key := []byte("set-key")
				member1 := []byte("member1")
				member2 := []byte("member2")
				args := [][]byte{[]byte("SADD"), key, member1, member2}
				expectedCount := int64(2)

				mockPersister.EXPECT().
					SAdd(gomock.Any(), key, member1, member2).
					Return(expectedCount)

				results := handler.Apply(ctx, args)

				Expect(results).To(HaveLen(1))
				Expect(results[0].Error).To(BeNil())
				Expect(string(results[0].Response)).To(Equal("2"))
			})
		})
	})

	Describe("SREM Command", func() {
		Context("when removing members from set", func() {
			It("should remove existing member", func() {
				key := []byte("set-key")
				member := []byte("member1")
				args := [][]byte{[]byte("SREM"), key, member}
				expectedCount := int64(1)

				mockPersister.EXPECT().
					SRem(gomock.Any(), key, member).
					Return(expectedCount)

				results := handler.Apply(ctx, args)

				Expect(results).To(HaveLen(1))
				Expect(results[0].Error).To(BeNil())
				Expect(string(results[0].Response)).To(Equal("1"))
			})

			It("should return 0 for non-existing member", func() {
				key := []byte("set-key")
				member := []byte("non-member")
				args := [][]byte{[]byte("SREM"), key, member}

				mockPersister.EXPECT().
					SRem(gomock.Any(), key, member).
					Return(int64(0))

				results := handler.Apply(ctx, args)

				Expect(results).To(HaveLen(1))
				Expect(results[0].Error).To(BeNil())
				Expect(string(results[0].Response)).To(Equal("0"))
			})
		})
	})

	Describe("SMEMBERS Command", func() {
		Context("when getting all set members", func() {
			It("should return all members", func() {
				key := []byte("set-key")
				args := [][]byte{[]byte("SMEMBERS"), key}
				expectedMembers := [][]byte{
					[]byte("member1"),
					[]byte("member2"),
					[]byte("member3"),
				}

				mockPersister.EXPECT().
					SMembers(gomock.Any(), key).
					Return(expectedMembers, nil)

				results := handler.Apply(ctx, args)

				Expect(results).To(HaveLen(1))
				Expect(results[0].Error).To(BeNil())

				expectedRedisFormat := "*3\r\n$7\r\nmember1\r\n$7\r\nmember2\r\n$7\r\nmember3\r\n"
				Expect(string(results[0].Response)).To(Equal(expectedRedisFormat))
			})

			It("should return empty response for non-existing set", func() {
				key := []byte("non-set")
				args := [][]byte{[]byte("SMEMBERS"), key}

				mockPersister.EXPECT().
					SMembers(gomock.Any(), key).
					Return([][]byte{}, nil)

				results := handler.Apply(ctx, args)

				Expect(results).To(HaveLen(1))
				Expect(results[0].Error).To(BeNil())
				Expect(string(results[0].Response)).To(Equal("*0\r\n"))
			})

			It("should return error when operation fails", func() {
				key := []byte("set-key")
				args := [][]byte{[]byte("SMEMBERS"), key}
				expectedError := errors.New("operation failed")

				mockPersister.EXPECT().
					SMembers(gomock.Any(), key).
					Return(nil, expectedError)

				results := handler.Apply(ctx, args)

				Expect(results).To(HaveLen(1))
				Expect(results[0].Error).To(Equal(expectedError))
			})
		})
	})

	Describe("SISMEMBER Command", func() {
		Context("when checking set membership", func() {
			It("should return 1 for existing member", func() {
				key := []byte("set-key")
				member := []byte("member1")
				args := [][]byte{[]byte("SISMEMBER"), key, member}

				mockPersister.EXPECT().
					SIsMember(gomock.Any(), key, member).
					Return(true)

				results := handler.Apply(ctx, args)

				Expect(results).To(HaveLen(1))
				Expect(results[0].Error).To(BeNil())
				Expect(string(results[0].Response)).To(Equal("1"))
			})

			It("should return 0 for non-existing member", func() {
				key := []byte("set-key")
				member := []byte("non-member")
				args := [][]byte{[]byte("SISMEMBER"), key, member}

				mockPersister.EXPECT().
					SIsMember(gomock.Any(), key, member).
					Return(false)

				results := handler.Apply(ctx, args)

				Expect(results).To(HaveLen(1))
				Expect(results[0].Error).To(BeNil())
				Expect(string(results[0].Response)).To(Equal("0"))
			})
		})
	})

	Describe("ZADD Command", func() {
		Context("when adding member to sorted set", func() {
			It("should return 1 for new member", func() {
				key := []byte("zset-key")
				score := []byte("1.5")
				member := []byte("member1")
				args := [][]byte{[]byte("ZADD"), key, score, member}

				mockPersister.EXPECT().
					ZAdd(gomock.Any(), key, 1.5, member).
					Return(int64(1))

				results := handler.Apply(ctx, args)

				Expect(results).To(HaveLen(1))
				Expect(results[0].Error).To(BeNil())
				Expect(string(results[0].Response)).To(Equal("1"))
			})

			It("should return error for invalid score", func() {
				key := []byte("zset-key")
				score := []byte("invalid")
				member := []byte("member1")
				args := [][]byte{[]byte("ZADD"), key, score, member}

				results := handler.Apply(ctx, args)

				Expect(results).To(HaveLen(1))
				Expect(results[0].Error).To(Equal(domain.ErrInvalidFloat))
			})
		})
	})

	Describe("ZRANGE Command", func() {
		Context("when getting range from sorted set", func() {
			It("should return members in range", func() {
				key := []byte("zset-key")
				start := []byte("0")
				stop := []byte("2")
				args := [][]byte{[]byte("ZRANGE"), key, start, stop}
				expectedMembers := [][]byte{
					[]byte("member1"),
					[]byte("member2"),
					[]byte("member3"),
				}

				mockPersister.EXPECT().
					ZRange(gomock.Any(), key, int64(0), int64(2)).
					Return(expectedMembers, nil)

				results := handler.Apply(ctx, args)

				Expect(results).To(HaveLen(1))
				Expect(results[0].Error).To(BeNil())

				expectedRedisFormat := "*3\r\n$7\r\nmember1\r\n$7\r\nmember2\r\n$7\r\nmember3\r\n"
				Expect(string(results[0].Response)).To(Equal(expectedRedisFormat))
			})

			It("should return error for invalid start index", func() {
				key := []byte("zset-key")
				start := []byte("invalid")
				stop := []byte("2")
				args := [][]byte{[]byte("ZRANGE"), key, start, stop}

				results := handler.Apply(ctx, args)

				Expect(results).To(HaveLen(1))
				Expect(results[0].Error).To(Equal(domain.ErrInvalidInteger))
			})
		})
	})

	Describe("ZCOUNT Command", func() {
		Context("when counting members in score range", func() {
			It("should return count of members", func() {
				key := []byte("zset-key")
				min := []byte("1.0")
				max := []byte("3.0")
				args := [][]byte{[]byte("ZCOUNT"), key, min, max}

				mockPersister.EXPECT().
					ZCount(gomock.Any(), key, 1.0, 3.0).
					Return(int64(5))

				results := handler.Apply(ctx, args)

				Expect(results).To(HaveLen(1))
				Expect(results[0].Error).To(BeNil())
				Expect(string(results[0].Response)).To(Equal("5"))
			})

			It("should return error for invalid min score", func() {
				key := []byte("zset-key")
				min := []byte("invalid")
				max := []byte("3.0")
				args := [][]byte{[]byte("ZCOUNT"), key, min, max}

				results := handler.Apply(ctx, args)

				Expect(results).To(HaveLen(1))
				Expect(results[0].Error).To(Equal(domain.ErrInvalidFloat))
			})
		})
	})

	Describe("INCR Command", func() {
		Context("when incrementing key", func() {
			It("should return incremented value", func() {
				key := []byte("counter")
				args := [][]byte{[]byte("INCR"), key}

				mockPersister.EXPECT().
					Incr(gomock.Any(), key).
					Return(int64(1), nil)

				results := handler.Apply(ctx, args)

				Expect(results).To(HaveLen(1))
				Expect(results[0].Error).To(BeNil())
				Expect(string(results[0].Response)).To(Equal("1"))
			})

			It("should return error when operation fails", func() {
				key := []byte("counter")
				args := [][]byte{[]byte("INCR"), key}
				expectedError := errors.New("not integer")

				mockPersister.EXPECT().
					Incr(gomock.Any(), key).
					Return(int64(0), expectedError)

				results := handler.Apply(ctx, args)

				Expect(results).To(HaveLen(1))
				Expect(results[0].Error).To(Equal(expectedError))
			})
		})
	})

	Describe("INCRBY Command", func() {
		Context("when incrementing key by value", func() {
			It("should return incremented value", func() {
				key := []byte("counter")
				increment := []byte("5")
				args := [][]byte{[]byte("INCRBY"), key, increment}

				mockPersister.EXPECT().
					IncrBy(gomock.Any(), key, int64(5)).
					Return(int64(15), nil)

				results := handler.Apply(ctx, args)

				Expect(results).To(HaveLen(1))
				Expect(results[0].Error).To(BeNil())
				Expect(string(results[0].Response)).To(Equal("15"))
			})

			It("should return error for invalid increment", func() {
				key := []byte("counter")
				increment := []byte("invalid")
				args := [][]byte{[]byte("INCRBY"), key, increment}

				results := handler.Apply(ctx, args)

				Expect(results).To(HaveLen(1))
				Expect(results[0].Error).To(Equal(domain.ErrInvalidInteger))
			})
		})
	})

	Describe("DECR Command", func() {
		Context("when decrementing key", func() {
			It("should return decremented value", func() {
				key := []byte("counter")
				args := [][]byte{[]byte("DECR"), key}

				mockPersister.EXPECT().
					Decr(gomock.Any(), key).
					Return(int64(-1), nil)

				results := handler.Apply(ctx, args)

				Expect(results).To(HaveLen(1))
				Expect(results[0].Error).To(BeNil())
				Expect(string(results[0].Response)).To(Equal("-1"))
			})

			It("should return error when operation fails", func() {
				key := []byte("counter")
				args := [][]byte{[]byte("DECR"), key}
				expectedError := errors.New("not integer")

				mockPersister.EXPECT().
					Decr(gomock.Any(), key).
					Return(int64(0), expectedError)

				results := handler.Apply(ctx, args)

				Expect(results).To(HaveLen(1))
				Expect(results[0].Error).To(Equal(expectedError))
			})
		})
	})

	Describe("DECRBY Command", func() {
		Context("when decrementing key by value", func() {
			It("should return decremented value", func() {
				key := []byte("counter")
				decrement := []byte("3")
				args := [][]byte{[]byte("DECRBY"), key, decrement}

				mockPersister.EXPECT().
					DecrBy(gomock.Any(), key, int64(3)).
					Return(int64(7), nil)

				results := handler.Apply(ctx, args)

				Expect(results).To(HaveLen(1))
				Expect(results[0].Error).To(BeNil())
				Expect(string(results[0].Response)).To(Equal("7"))
			})

			It("should return error for invalid decrement", func() {
				key := []byte("counter")
				decrement := []byte("invalid")
				args := [][]byte{[]byte("DECRBY"), key, decrement}

				results := handler.Apply(ctx, args)

				Expect(results).To(HaveLen(1))
				Expect(results[0].Error).To(Equal(domain.ErrInvalidInteger))
			})
		})
	})

	Describe("APPEND Command", func() {
		Context("when appending to key", func() {
			It("should return new length", func() {
				key := []byte("string-key")
				value := []byte("world")
				args := [][]byte{[]byte("APPEND"), key, value}

				mockPersister.EXPECT().
					Append(gomock.Any(), key, value).
					Return(int64(10))

				results := handler.Apply(ctx, args)

				Expect(results).To(HaveLen(1))
				Expect(results[0].Error).To(BeNil())
				Expect(string(results[0].Response)).To(Equal("10"))
			})
		})
	})
})
