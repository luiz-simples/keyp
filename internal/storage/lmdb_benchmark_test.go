package storage_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/luiz-simples/keyp.git/internal/storage"
)

func BenchmarkLMDBStorage(b *testing.B) {
	tmpDir, err := os.MkdirTemp("", "lmdb-benchmark-*")
	if err != nil {
		b.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	lmdbStorage, err := storage.NewLMDBStorage(tmpDir)
	if err != nil {
		b.Fatal(err)
	}
	defer lmdbStorage.Close()

	b.Run("Set", func(b *testing.B) {
		benchmarkSet(b, lmdbStorage)
	})

	b.Run("Get", func(b *testing.B) {
		benchmarkGet(b)
	})

	b.Run("Del", func(b *testing.B) {
		benchmarkDel(b)
	})

	b.Run("SetGetDel", func(b *testing.B) {
		benchmarkSetGetDel(b, lmdbStorage)
	})
}

func benchmarkSet(b *testing.B, storage *storage.LMDBStorage) {
	key := []byte("benchmark-key")
	value := []byte("benchmark-value-with-some-data")

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		keyWithIndex := append(key, []byte(fmt.Sprintf("-%d", i))...)
		err := storage.Set(keyWithIndex, value)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func benchmarkGet(b *testing.B) {
	tmpDir, err := os.MkdirTemp("", "lmdb-benchmark-get-*")
	if err != nil {
		b.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	storage, err := storage.NewLMDBStorage(tmpDir)
	if err != nil {
		b.Fatal(err)
	}
	defer storage.Close()

	key := []byte("benchmark-key")
	value := []byte("benchmark-value-with-some-data")

	keys := make([][]byte, b.N)
	for i := 0; i < b.N; i++ {
		keyWithIndex := append(key, []byte(fmt.Sprintf("-%d", i))...)
		keys[i] = keyWithIndex
		err := storage.Set(keyWithIndex, value)
		if err != nil {
			b.Fatal(err)
		}
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, err := storage.Get(keys[i])
		if err != nil {
			b.Fatal(err)
		}
	}
}

func benchmarkDel(b *testing.B) {
	tmpDir, err := os.MkdirTemp("", "lmdb-benchmark-del-*")
	if err != nil {
		b.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	storage, err := storage.NewLMDBStorage(tmpDir)
	if err != nil {
		b.Fatal(err)
	}
	defer storage.Close()

	key := []byte("benchmark-key")
	value := []byte("benchmark-value-with-some-data")

	keys := make([][]byte, b.N)
	for i := 0; i < b.N; i++ {
		keyWithIndex := append(key, []byte(fmt.Sprintf("-%d", i))...)
		keys[i] = keyWithIndex
		err := storage.Set(keyWithIndex, value)
		if err != nil {
			b.Fatal(err)
		}
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, err := storage.Del(keys[i])
		if err != nil {
			b.Fatal(err)
		}
	}
}

func benchmarkSetGetDel(b *testing.B, storage *storage.LMDBStorage) {
	key := []byte("benchmark-key")
	value := []byte("benchmark-value-with-some-data")

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		keyWithIndex := append(key, []byte(fmt.Sprintf("-%d", i))...)

		err := storage.Set(keyWithIndex, value)
		if err != nil {
			b.Fatal(err)
		}

		_, err = storage.Get(keyWithIndex)
		if err != nil {
			b.Fatal(err)
		}

		_, err = storage.Del(keyWithIndex)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkLMDBStorageVariousSizes(b *testing.B) {
	tmpDir, err := os.MkdirTemp("", "lmdb-benchmark-sizes-*")
	if err != nil {
		b.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	storage, err := storage.NewLMDBStorage(tmpDir)
	if err != nil {
		b.Fatal(err)
	}
	defer storage.Close()

	sizes := []int{10, 100, 1024, 10240, 102400}

	for _, size := range sizes {
		b.Run(fmt.Sprintf("ValueSize%d", size), func(b *testing.B) {
			benchmarkSetWithValueSize(b, storage, size)
		})
	}
}

func benchmarkSetWithValueSize(b *testing.B, storage *storage.LMDBStorage, valueSize int) {
	key := []byte("benchmark-key")
	value := make([]byte, valueSize)
	for i := range value {
		value[i] = byte(i % 256)
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		keyWithIndex := append(key, []byte(fmt.Sprintf("-%d", i))...)
		err := storage.Set(keyWithIndex, value)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkLMDBStorageConcurrent(b *testing.B) {
	tmpDir, err := os.MkdirTemp("", "lmdb-benchmark-concurrent-*")
	if err != nil {
		b.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	storage, err := storage.NewLMDBStorage(tmpDir)
	if err != nil {
		b.Fatal(err)
	}
	defer storage.Close()

	value := []byte("benchmark-value-concurrent")

	b.ResetTimer()
	b.ReportAllocs()

	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			key := []byte(fmt.Sprintf("concurrent-key-%d", i))
			err := storage.Set(key, value)
			if err != nil {
				b.Fatal(err)
			}
			i++
		}
	})
}
