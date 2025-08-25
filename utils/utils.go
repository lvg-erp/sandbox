package utils

import (
	"context"
	rand_m "crypto/rand"
	"errors"
	"fmt"
	"golang.org/x/exp/rand"
	"hash/fnv"
	"log"
	"math/big"
	"time"
)

const (
	low    = "03"
	medium = "02"
	high   = "01"
)

func GenerateString() string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	const length = 6
	rand.Seed(uint64(time.Now().UnixNano()))

	b := make([]byte, length)
	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}
	return string(b)

}

func GenerateRandomBytes() []byte {
	fixedSet := []string{
		"K9xP4m",
		"tR7wQ2",
		"Z8nL5j",
		"aB1cD2",
		"Xy9zW4",
		"Pq3Rs5",
		"Lm8Nk7",
		"Jh2Tv9",
		"Fg6Yx1",
		"De4Uw8",
	}
	rand.Seed(uint64(time.Now().UnixNano()))
	return []byte(fixedSet[rand.Intn(len(fixedSet))])
}

func GenerateUniqueRandomNumbers(n, min, max int, seed int64) ([]int, error) {
	if n > max-min+1 {
		return nil, errors.New("requested number of unique values exceeds range")
	}

	numbers := make([]int, max-min+1)
	for i := range numbers {
		numbers[i] = min + i
	}

	rng := rand.New(rand.NewSource(uint64(seed)))
	for i := len(numbers) - 1; i > 0; i-- {
		j := rng.Intn(i + 1)
		numbers[i], numbers[j] = numbers[j], numbers[i]
	}

	return numbers[:n], nil
}

func RandomFomConstants() {
	rand.Seed(uint64(time.Now().UnixNano()))

	// Срез с константами
	priorities := []string{low, medium, high}

	// Случайный выбор
	randomPriority := priorities[rand.Intn(len(priorities))]

	fmt.Printf("Random priority: %s\n", randomPriority)
}

func GenRandP() int {
	randomIndex, err := rand_m.Int(rand_m.Reader, big.NewInt(100))
	if err != nil {
		log.Fatalf("Failed to generate random number: %v", err)
	}
	return int(randomIndex.Int64())
}

func Contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

func StringHash(s string) uint32 {
	h := fnv.New32a()
	h.Write([]byte(s))
	return h.Sum32()
}

// факториал числа
func factorialInt(n int) *big.Int {
	result := big.NewInt(1)
	for i := 2; i <= n; i++ {
		result.Mul(result, big.NewInt(int64(i)))
	}
	return result
}

// факториал с контекстом
func factorialInt_(n int, ctx context.Context) (*big.Int, error) {
	if n < 0 {
		return nil, fmt.Errorf("negative number %d", n)
	}
	result := big.NewInt(1)
	for i := 2; i <= n; i++ {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
			result.Mul(result, big.NewInt(int64(i)))
		}
	}
	if !result.IsInt64() {
		return nil, fmt.Errorf("factorial %d too large for int64", n)
	}
	return result, nil
}
