package models

import (
	"context"
	"encoding/json"
	"github.com/redis/go-redis/v9"
	"rediskafkago/config"
	"sync"
)

type DataType interface {
	Address | Label | Transaction
}

type MetadataDB[T DataType] struct {
	db int
	mu sync.RWMutex
}

var client = redis.NewClient(&redis.Options{
	Addr:     config.RedisServerAddr,
	Password: "", // no password set
	DB:       0,  // use default DB
})

func NewMetadataDB[T DataType](db int) *MetadataDB[T] {
	return &MetadataDB[T]{
		db: db,
	}
}

func (m *MetadataDB[T]) KeyExists(key string) (bool, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	ctx := context.Background()
	//выбререм базу данных
	if _, err := client.Do(ctx, "SELECT", m.db).Result(); err != nil {
		return false, err
	}

	exists, err := client.Exists(ctx, key).Result()
	if err != nil {
		return false, err
	}

	return exists == 1, nil
}

func (m *MetadataDB[T]) Get(key string) (T, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	var res T
	ctx := context.Background()
	if _, err := client.Do(ctx, "SELECT", m.db).Result(); err != nil {
		return res, err
	}

	val, err := client.Get(ctx, key).Result()

	if err != nil {
		return res, err
	}
	err = json.Unmarshal([]byte(val), &res)

	if err != nil {
		return res, err
	}

	return res, nil

}

func (m *MetadataDB[T]) GetAll(client *redis.Client) ([]T, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var res []T
	ctx := context.Background()

	//выберем базу данных
	if _, err := client.Do(ctx, "SELECT", m.db).Result(); err != nil {
		return nil, err
	}

	//заберем все ключи из базы
	keys, err := client.Keys(ctx, "*").Result()
	if err != nil {
		return nil, err
	}

	for _, key := range keys {
		value, err := client.Get(ctx, key).Result()
		if err != nil {
			return nil, err
		}

		var data T
		err = json.Unmarshal([]byte(value), &data)
		if err != nil {
			return nil, err
		}
		res = append(res, data)
	}
	return res, nil
}

func (m *MetadataDB[T]) Update(key string, data T) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	ctx := context.Background()
	if _, err := client.Do(ctx, "SELECT", m.db).Result(); err != nil {
		return err
	}

	dataJSON, err := json.Marshal(data)
	if err != nil {
		return err
	}
	err = client.Set(ctx, key, dataJSON, 0).Err()
	if err != nil {
		return err
	}

	return nil
}

func (m *MetadataDB[T]) Delete(key string) (int64, error) {

	m.mu.Lock()
	defer m.mu.Unlock()
	ctx := context.Background()
	if _, err := client.Do(ctx, "SELECT", m.db).Result(); err != nil {
		return 0, err
	}

	keyDeleted, err := client.Del(ctx, key).Result()
	if err != nil {
		return 0, err
	}

	return keyDeleted, nil
}

var AddressDB = NewMetadataDB[Address](0)
var LabelDB = NewMetadataDB[Label](1)
var TransactionDB = NewMetadataDB[Transaction](2)
