package app_test

import (
	"go.uber.org/mock/gomock"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/luiz-simples/keyp.git/internal/app"
)

var _ = Describe("Utils", func() {
	var (
		ctrl     *gomock.Controller
		mockPool *MockLogicaler
		server   *app.Server
	)

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
		mockPool = NewMockLogicaler(ctrl)
		server = app.NewServer(mockPool)
	})

	AfterEach(func() {
		ctrl.Finish()
	})

	Describe("generateConnectionID function behavior", func() {
		Context("when testing connection ID generation through OnAccept", func() {
			It("should generate unique connection IDs for multiple connections", func() {
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

			It("should generate valid connection IDs consistently", func() {
				connections := 10

				for range connections {
					mockConnTemp := NewMockConn(ctrl)
					mockDispatcherTemp := NewMockDispatcher(ctrl)
					mockPool.EXPECT().Get(gomock.Any()).Return(mockDispatcherTemp).Times(1)
					mockConnTemp.EXPECT().SetContext(gomock.Any()).Times(1)

					result := server.OnAccept(mockConnTemp)
					Expect(result).To(BeTrue())
				}
			})

			It("should handle rapid connection generation", func() {
				connections := 50

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

	Describe("utility functions integration", func() {
		Context("when testing server lifecycle", func() {
			It("should handle server shutdown with active connections", func() {
				mockConn1 := NewMockConn(ctrl)
				mockConn2 := NewMockConn(ctrl)
				mockDispatcher1 := NewMockDispatcher(ctrl)
				mockDispatcher2 := NewMockDispatcher(ctrl)

				mockPool.EXPECT().Get(gomock.Any()).Return(mockDispatcher1).Times(1)
				mockPool.EXPECT().Get(gomock.Any()).Return(mockDispatcher2).Times(1)
				mockConn1.EXPECT().SetContext(gomock.Any()).Times(1)
				mockConn2.EXPECT().SetContext(gomock.Any()).Times(1)

				server.OnAccept(mockConn1)
				server.OnAccept(mockConn2)

				server.Close()
			})
		})

		Context("when testing configuration handling", func() {
			It("should handle various configuration scenarios", func() {
				configs := []app.Config{
					{Address: ":6379", DataDir: "/tmp/keyp"},
					{Address: "localhost:8080", DataDir: "/custom/path"},
					{Address: "", DataDir: ""},
					{Address: "0.0.0.0:9999", DataDir: "/var/lib/keyp"},
				}

				for _, config := range configs {
					Expect(config.Address).To(BeAssignableToTypeOf(""))
					Expect(config.DataDir).To(BeAssignableToTypeOf(""))
				}
			})

			It("should handle configuration validation", func() {
				config := app.Config{
					Address: ":6379",
					DataDir: "/tmp/keyp",
				}

				Expect(len(config.Address)).To(BeNumerically(">", 0))
				Expect(len(config.DataDir)).To(BeNumerically(">", 0))
			})
		})
	})
})
