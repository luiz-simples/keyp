package app_test

import (
	"context"

	"go.uber.org/mock/gomock"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/luiz-simples/keyp.git/internal/app"
)

var _ = Describe("Server Additional Tests", func() {
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

	Describe("NewServer", func() {
		It("should create a new server with the provided pool", func() {
			newServer := app.NewServer(mockPool)
			Expect(newServer).NotTo(BeNil())
		})
	})

	Describe("OnAccept", func() {
		It("should handle connection acceptance", func() {
			mockConn := NewMockConn(ctrl)

			// Mock the connection context setup
			mockConn.EXPECT().SetContext(gomock.Any()).Return()

			// Mock pool operations
			mockDispatcher := NewMockDispatcher(ctrl)
			mockPool.EXPECT().Get(gomock.Any()).Return(mockDispatcher)

			server.OnAccept(mockConn)
		})
	})

	Describe("OnClosed", func() {
		It("should handle connection closure with invalid context", func() {
			mockConn := NewMockConn(ctrl)

			// Return invalid context (not context.Context)
			mockConn.EXPECT().Context().Return("invalid")

			server.OnClosed(mockConn, nil)
		})

		It("should handle connection closure with missing connection ID", func() {
			mockConn := NewMockConn(ctrl)

			// Return context without connection ID
			ctx := context.Background()
			mockConn.EXPECT().Context().Return(ctx)

			server.OnClosed(mockConn, nil)
		})
	})

	Describe("Config", func() {
		It("should create config with address", func() {
			config := app.Config{
				Address: ":6379",
			}
			Expect(config.Address).To(Equal(":6379"))
		})
	})
})
