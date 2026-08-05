package main

import (
	"cachetable/api"
	"fmt"
	"time"
)

func main() {
	fmt.Println("Запуск демо кэша...")

	// Создаем кэш, который будет чистить мусор каждую секунду
	c := api.NewCache(1 * time.Second)
	defer c.Close()

	// Записываем данные
	c.Set("session:123", "user_data", 3*time.Second)
	c.Set("config:db", "localhost:5432", 0) // 0 = бесконечно

	// Читаем
	if val, ok := c.Get("session:123"); ok {
		fmt.Println("Найдено:", val)
	}

	// Ждем протухания
	fmt.Println("Ждем 4 секунды...")
	time.Sleep(4 * time.Second)

	// Проверяем Lazy Expiration
	if _, ok := c.Get("session:123"); !ok {
		fmt.Println("Ключ session:123 протух и был удален при Get (Lazy)!")
	}

	// Проверяем, что бесконечный ключ жив
	if val, ok := c.Get("config:db"); ok {
		fmt.Println("Бесконечный ключ жив:", val)
	}
}
