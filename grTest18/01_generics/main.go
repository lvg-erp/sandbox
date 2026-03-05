package main

import (
	"context"
	"fmt"
	"sandbox/grTest18/01_generics/cacheUse"
	"time"
)

func main() {
	// Загрузчик
	loader := func(ctx context.Context, keys []string) (map[string]string, error) {
		result := make(map[string]string)
		for _, key := range keys {
			time.Sleep(10 * time.Millisecond)
			result[key] = "Record_" + key
		}

		return result, nil
	}

	// Кеш
	cache := cacheUse.NewCache[string, string](loader)

	// Данные
	ctx := context.Background()

	val, err := cache.Get(ctx, "ключ1")
	if err != nil {
		fmt.Println("Ошибка", err)
	} else {
		fmt.Println("Ключ1", val)
	}
	// Получаем несколько ключей
	keys := []string{"ключ2", "ключ3", "ключ4"}
	values, err := cache.GetMany(ctx, keys)
	if err != nil {
		fmt.Println("Ошибка", err)
	} else {
		fmt.Println("Результаты", values)
	}
	// Повторный вызов кэша
	val2, err := cache.Get(ctx, "ключ2")
	if err != nil {
		fmt.Println("Ошибка", err)
	} else {
		fmt.Println("Ключ2", val2)
	}

}
