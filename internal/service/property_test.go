package service_test

import (
	"context"
	"fmt"
	"strconv"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/luiz-simples/keyp.git/internal/service"
	"github.com/luiz-simples/keyp.git/internal/storage"
)

var _ = Describe("Handler Property-Based Tests", func() {
	var (
		handler     *service.Handler
		storageImpl *storage.Client
		ctx         context.Context
		properties  *gopter.Properties
		testDir     string
	)

	BeforeEach(func() {
		ctx = context.Background()

		testDir = createUniqueTestDir("property")

		var err error
		storageImpl, err = storage.NewClient(testDir)
		Expect(err).NotTo(HaveOccurred())
		handler = service.NewHandler(storageImpl)

		parameters := gopter.DefaultTestParameters()
		parameters.MinSuccessfulTests = 100
		parameters.MaxSize = 50
		properties = gopter.NewProperties(parameters)
	})

	AfterEach(func() {
		if storageImpl != nil {
			storageImpl.Close()
		}
		cleanupTestDir(testDir)
	})

	Describe("SET-GET Property", func() {
		It("should satisfy: SET(k,v) then GET(k) returns v", func() {
			property := prop.ForAll(
				func(key, value string) bool {
					if key == "" {
						return true
					}

					keyBytes := []byte(key)
					valueBytes := []byte(value)

					setArgs := [][]byte{[]byte("SET"), keyBytes, valueBytes}
					setResults := handler.Apply(ctx, setArgs)

					if len(setResults) != 1 || setResults[0].Error != nil {
						return false
					}

					getArgs := [][]byte{[]byte("GET"), keyBytes}
					getResults := handler.Apply(ctx, getArgs)

					if len(getResults) != 1 || getResults[0].Error != nil {
						return false
					}

					return string(getResults[0].Response) == value
				},
				gen.AlphaString().SuchThat(func(s string) bool { return len(s) > 0 && len(s) < 100 }),
				gen.AlphaString().SuchThat(func(s string) bool { return len(s) < 1000 }),
			)

			properties.Property("SET-GET consistency", property)
			Expect(properties.Run(gopter.ConsoleReporter(false))).To(BeTrue())
		})
	})

	Describe("DEL Property", func() {
		It("should satisfy: SET(k,v) then DEL(k) then GET(k) returns nil", func() {
			property := prop.ForAll(
				func(key, value string) bool {
					if key == "" {
						return true
					}

					keyBytes := []byte(key)
					valueBytes := []byte(value)

					setArgs := [][]byte{[]byte("SET"), keyBytes, valueBytes}
					setResults := handler.Apply(ctx, setArgs)

					if len(setResults) != 1 || setResults[0].Error != nil {
						return false
					}

					delArgs := [][]byte{[]byte("DEL"), keyBytes}
					delResults := handler.Apply(ctx, delArgs)

					if len(delResults) != 1 || delResults[0].Error != nil {
						return false
					}

					deletedCount := string(delResults[0].Response)
					if deletedCount != "1" {
						return false
					}

					getArgs := [][]byte{[]byte("GET"), keyBytes}
					getResults := handler.Apply(ctx, getArgs)

					if len(getResults) != 1 || getResults[0].Error != nil {
						return false
					}

					return getResults[0].Response == nil
				},
				gen.AlphaString().SuchThat(func(s string) bool { return len(s) > 0 && len(s) < 100 }),
				gen.AlphaString().SuchThat(func(s string) bool { return len(s) < 1000 }),
			)

			properties.Property("SET-DEL-GET consistency", property)
			Expect(properties.Run(gopter.ConsoleReporter(false))).To(BeTrue())
		})
	})

	Describe("Multiple DEL Property", func() {
		It("should satisfy: DEL count equals number of existing keys", func() {
			property := prop.ForAll(
				func(keys []string) bool {
					if len(keys) == 0 {
						return true
					}

					uniqueKeys := make(map[string]bool)
					validKeys := make([]string, 0)
					for _, key := range keys {
						if key != "" && !uniqueKeys[key] {
							uniqueKeys[key] = true
							validKeys = append(validKeys, key)
						}
					}

					if len(validKeys) == 0 {
						return true
					}

					for i, key := range validKeys {
						value := fmt.Sprintf("value_%d", i)
						setArgs := [][]byte{[]byte("SET"), []byte(key), []byte(value)}
						setResults := handler.Apply(ctx, setArgs)

						if len(setResults) != 1 || setResults[0].Error != nil {
							return false
						}
					}

					delArgs := make([][]byte, len(validKeys)+1)
					delArgs[0] = []byte("DEL")
					for i, key := range validKeys {
						delArgs[i+1] = []byte(key)
					}

					delResults := handler.Apply(ctx, delArgs)

					if len(delResults) != 1 || delResults[0].Error != nil {
						return false
					}

					deletedCount := string(delResults[0].Response)
					return deletedCount == strconv.Itoa(len(validKeys))
				},
				gen.SliceOf(gen.AlphaString().SuchThat(func(s string) bool { return len(s) < 50 })).
					SuchThat(func(slice []string) bool { return len(slice) <= 10 }),
			)

			properties.Property("Multiple DEL count consistency", property)
			Expect(properties.Run(gopter.ConsoleReporter(false))).To(BeTrue())
		})
	})

	Describe("Command Case Insensitivity Property", func() {
		It("should satisfy: commands work regardless of case", func() {
			property := prop.ForAll(
				func(key, value string) bool {
					if key == "" {
						return true
					}

					keyBytes := []byte(key)
					valueBytes := []byte(value)

					commands := []string{"SET", "set", "Set", "sEt"}

					for _, cmd := range commands {
						setArgs := [][]byte{[]byte(cmd), keyBytes, valueBytes}
						setResults := handler.Apply(ctx, setArgs)

						if len(setResults) != 1 || setResults[0].Error != nil {
							return false
						}

						getArgs := [][]byte{[]byte("GET"), keyBytes}
						getResults := handler.Apply(ctx, getArgs)

						if len(getResults) != 1 || getResults[0].Error != nil {
							return false
						}

						if string(getResults[0].Response) != value {
							return false
						}
					}

					return true
				},
				gen.AlphaString().SuchThat(func(s string) bool { return len(s) > 0 && len(s) < 50 }),
				gen.AlphaString().SuchThat(func(s string) bool { return len(s) < 100 }),
			)

			properties.Property("Command case insensitivity", property)
			Expect(properties.Run(gopter.ConsoleReporter(false))).To(BeTrue())
		})
	})

	Describe("PING Property", func() {
		It("should satisfy: PING always returns PONG or echoes message", func() {
			property := prop.ForAll(
				func(message string) bool {
					var args [][]byte

					args = [][]byte{[]byte("PING")}

					if message != "" {
						args = [][]byte{[]byte("PING"), []byte(message)}
					}

					results := handler.Apply(ctx, args)

					if len(results) != 1 || results[0].Error != nil {
						return false
					}

					if message == "" {
						return string(results[0].Response) == "PONG"
					}

					return string(results[0].Response) == message
				},
				gen.AlphaString().SuchThat(func(s string) bool { return len(s) < 100 }),
			)

			properties.Property("PING response consistency", property)
			Expect(properties.Run(gopter.ConsoleReporter(false))).To(BeTrue())
		})
	})

	Describe("Argument Validation Property", func() {
		It("should satisfy: invalid argument counts always return errors", func() {
			property := prop.ForAll(
				func(argCount int) bool {
					if argCount < 0 || argCount > 20 {
						return true
					}

					args := make([][]byte, argCount)
					for i := range argCount {
						args[i] = []byte(fmt.Sprintf("arg%d", i))
					}

					if argCount > 0 {
						args[0] = []byte("SET")
					}

					results := handler.Apply(ctx, args)

					if len(results) != 1 {
						return false
					}

					if argCount >= 3 && argCount <= 5 {
						return results[0].Error == nil
					}

					return results[0].Error != nil
				},
				gen.IntRange(0, 10),
			)

			properties.Property("Argument validation consistency", property)
			Expect(properties.Run(gopter.ConsoleReporter(false))).To(BeTrue())
		})
	})
})
