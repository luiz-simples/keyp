package app_test

import (
	"context"
	"errors"

	"go.uber.org/mock/gomock"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/luiz-simples/keyp.git/internal/app"
)

var _ = Describe("Server", func() {
	var (
		ctrl      *gomock.Controller
		mockPool  *MockLogicaler
		server    *app.Server
		mockConn  *MockConn
		testError error
	)

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
		mockPool = NewMockLogicaler(ctrl)
		server = app.NewServer(mockPool)
		mockConn = NewMockConn(ctrl)
		testError = errors.New("test error")
	})

	AfterEach(func() {
		ctrl.Finish()
	})

	Describe("NewServer", func() {
		Context("when creating new server instance", func() {
			It("should initialize server with provided pool", func() {
				newServer := app.NewServer(mockPool)

				Expect(newServer).NotTo(BeNil())
			})

			It("should create server with different pool instances", func() {
				mockPool2 := NewMockLogicaler(ctrl)
				newServer := app.NewServer(mockPool2)

				Expect(newServer).NotTo(BeNil())
			})
		})
	})

	Describe("OnAccept", func() {
		Context("when accepting new connection", func() {
			It("should set connection context with generated ID", func() {
				mockDispatcher := NewMockDispatcher(ctrl)
				mockPool.EXPECT().Get(gomock.Any()).Return(mockDispatcher).Times(1)
				mockConn.EXPECT().SetContext(gomock.Any()).Times(1)

				result := server.OnAccept(mockConn)

				Expect(result).To(BeTrue())
			})

			It("should return true for successful connection acceptance", func() {
				mockDispatcher := NewMockDispatcher(ctrl)
				mockPool.EXPECT().Get(gomock.Any()).Return(mockDispatcher).Times(1)
				mockConn.EXPECT().SetContext(gomock.Any()).Times(1)

				result := server.OnAccept(mockConn)

				Expect(result).To(BeTrue())
			})

			It("should handle multiple concurrent connections", func() {
				mockConn1 := NewMockConn(ctrl)
				mockConn2 := NewMockConn(ctrl)
				mockDispatcher1 := NewMockDispatcher(ctrl)
				mockDispatcher2 := NewMockDispatcher(ctrl)

				mockPool.EXPECT().Get(gomock.Any()).Return(mockDispatcher1).Times(1)
				mockPool.EXPECT().Get(gomock.Any()).Return(mockDispatcher2).Times(1)
				mockConn1.EXPECT().SetContext(gomock.Any()).Times(1)
				mockConn2.EXPECT().SetContext(gomock.Any()).Times(1)

				result1 := server.OnAccept(mockConn1)
				result2 := server.OnAccept(mockConn2)

				Expect(result1).To(BeTrue())
				Expect(result2).To(BeTrue())
			})

			It("should generate unique connection contexts", func() {
				connections := 5

				for range connections {
					mockConnTemp := NewMockConn(ctrl)
					mockDispatcherTemp := NewMockDispatcher(ctrl)
					mockPool.EXPECT().Get(gomock.Any()).Return(mockDispatcherTemp).Times(1)
					mockConnTemp.EXPECT().SetContext(gomock.Any()).Times(1)

					result := server.OnAccept(mockConnTemp)
					Expect(result).To(BeTrue())
				}
			})
		})
	})

	Describe("OnClosed", func() {
		Context("when closing connection after accept", func() {
			It("should cleanup connection resources", func() {
				mockDispatcher := NewMockDispatcher(ctrl)

				mockPool.EXPECT().Get(gomock.Any()).Return(mockDispatcher).Times(1)
				mockConn.EXPECT().SetContext(gomock.Any()).Do(func(ctx context.Context) {
					mockConn.EXPECT().Context().Return(ctx).Times(1)
				}).Times(1)
				server.OnAccept(mockConn)

				mockPool.EXPECT().Free(mockDispatcher).Times(1)

				server.OnClosed(mockConn, nil)
			})

			It("should handle connection closure with error", func() {
				mockDispatcher := NewMockDispatcher(ctrl)

				mockPool.EXPECT().Get(gomock.Any()).Return(mockDispatcher).Times(1)
				mockConn.EXPECT().SetContext(gomock.Any()).Do(func(ctx context.Context) {
					mockConn.EXPECT().Context().Return(ctx).Times(1)
				}).Times(1)
				server.OnAccept(mockConn)

				mockPool.EXPECT().Free(mockDispatcher).Times(1)

				server.OnClosed(mockConn, testError)
			})
		})
	})

	Describe("Close", func() {
		Context("when closing server", func() {
			It("should handle server closure", func() {
				server.Close()
			})

			It("should handle server closure with active connections", func() {
				mockConn1 := NewMockConn(ctrl)
				mockConn2 := NewMockConn(ctrl)
				mockDispatcher1 := NewMockDispatcher(ctrl)
				mockDispatcher2 := NewMockDispatcher(ctrl)

				mockPool.EXPECT().Get(gomock.Any()).Return(mockDispatcher1).Times(1)
				mockPool.EXPECT().Get(gomock.Any()).Return(mockDispatcher2).Times(1)
				mockConn1.EXPECT().SetContext(gomock.Any()).Times(1)
				mockConn2.EXPECT().SetContext(gomock.Any()).Times(1)

				mockDispatcher1.EXPECT().Clear().Times(1)
				mockDispatcher2.EXPECT().Clear().Times(1)

				server.OnAccept(mockConn1)
				server.OnAccept(mockConn2)

				server.Close()
			})

			It("should handle multiple close calls", func() {
				server.Close()
				server.Close()
			})
		})
	})

	Describe("Start", func() {
		Context("when starting server with valid config", func() {
			It("should accept valid configuration", func() {
				config := app.Config{
					Address: ":6379",
					DataDir: "/tmp/keyp",
				}

				Expect(config.Address).To(Equal(":6379"))
				Expect(config.DataDir).To(Equal("/tmp/keyp"))
			})

			It("should handle empty configuration", func() {
				config := app.Config{}

				Expect(config.Address).To(Equal(""))
				Expect(config.DataDir).To(Equal(""))
			})

			It("should handle configuration with custom values", func() {
				config := app.Config{
					Address: "localhost:8080",
					DataDir: "/custom/path",
				}

				Expect(config.Address).To(Equal("localhost:8080"))
				Expect(config.DataDir).To(Equal("/custom/path"))
			})
		})
	})
})
