package controllers

import (
	"encoding/json"
	"fmt"
	"github.com/IBM/sarama"
	"github.com/go-playground/validator"
	"github.com/labstack/echo/v4"
	"github.com/redis/go-redis/v9"
	"net/http"
	"rediskafkago/config"
	"rediskafkago/models"
)

var client = redis.NewClient(&redis.Options{
	Addr:     config.RedisServerAddr,
	Password: "",
	DB:       0,
})

// Получим адрес
func GetAddress(db *models.MetadataDB[models.Address]) echo.HandlerFunc {
	return func(c echo.Context) error {
		address := c.Param("address")
		data, err := db.Get(address)
		if err != nil {
			if err == redis.Nil {
				return c.JSON(http.StatusNotFound, "The requested address does not exist")
			} else {
				return c.JSON(http.StatusInternalServerError, err.Error())
			}
		}
		return c.JSON(http.StatusOK, data)
	}
}

// Получим все адреса
func GetAddresses(db *models.MetadataDB[models.Address]) echo.HandlerFunc {
	return func(c echo.Context) error {
		values, err := db.GetAll(client)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, err.Error())
		}
		return c.JSON(http.StatusOK, values)
	}
}

var validate = validator.New()

// POST `/addresses/{address}`
// создаем адрес
func CreateAddress(producer sarama.SyncProducer, db *models.MetadataDB[models.Address]) echo.HandlerFunc {
	return func(c echo.Context) error {
		address := c.Param("address")
		exists, err := db.KeyExists(address)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, err.Error())
		}

		if exists {
			return c.JSON(http.StatusBadRequest, "The resource to updater does not exist")
		}
		var data models.Address

		if err := c.Bind(&data); err != nil {
			return c.JSON(http.StatusBadRequest, err.Error())
		}

		if validationErr := validate.Struct(&data); validationErr != nil {
			return c.JSON(http.StatusBadRequest, validationErr.Error())
		}

		if address != data.Address {
			return c.JSON(http.StatusBadRequest, "address not match")
		}
		//создаем адрес

		newAddr := models.Address{
			Address: data.Address,
			Labels:  data.Labels,
		}
		// Отсылаем в очередь kafka
		newAddrJSON, err := json.Marshal(newAddr)
		if err != nil {
			return fmt.Errorf("failed to marshall Address: %w", err)
		}

		msg := &sarama.ProducerMessage{
			Topic: "address",
			Key:   sarama.StringEncoder(address),
			Value: sarama.StringEncoder(newAddrJSON),
		}

		if _, _, err = producer.SendMessage(msg); err != nil {
			return c.JSON(http.StatusInternalServerError, err.Error())
		}

		return c.JSON(http.StatusOK, "address updated")
	}

}

// PUT `/addresses/{address}`
// обновить адрес
func UpdateAddress(producer sarama.SyncProducer, db *models.MetadataDB[models.Address]) echo.HandlerFunc {
	return func(c echo.Context) error {
		address := c.Param("address")
		exists, err := db.KeyExists(address)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, err.Error())
		}
		if !exists {
			return c.JSON(http.StatusBadRequest, "The resource to update does not exist")
		}

		var data models.Address

		if err := c.Bind(&data); err != nil {
			return c.JSON(http.StatusBadRequest, err.Error())
		}
		if validationErr := validate.Struct(&data); validationErr != nil {
			return c.JSON(http.StatusBadRequest, validationErr.Error())
		}
		if address != data.Address {
			return c.JSON(http.StatusBadRequest, "address not match")
		}
		//
		newAddr := models.Address{
			Address: data.Address,
			Labels:  data.Labels,
		}

		//send kafka
		newAddrJSON, err := json.Marshal(newAddr)
		if err != nil {
			return fmt.Errorf("failed to marshall Address: %w", err)
		}

		msg := &sarama.ProducerMessage{
			Topic: "address",
			Key:   sarama.StringEncoder(address),
			Value: sarama.StringEncoder(newAddrJSON),
		}

		if _, _, err = producer.SendMessage(msg); err != nil {
			return c.JSON(http.StatusInternalServerError, err.Error())
		}

		return c.JSON(http.StatusOK, "address updated")
	}
}

// DELETE `/addresses/{address}`
// удалить адрес
func DeleteAddress(block sarama.SyncProducer, db *models.MetadataDB[models.Address]) echo.HandlerFunc {
	return func(c echo.Context) error {
		address := c.Param("address")
		keyDeleted, err := db.Delete(address)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, err.Error())
		}
		if keyDeleted == 0 {
			return c.JSON(http.StatusBadRequest, "The resouce to delete no exit")
		}

		return c.JSON(http.StatusOK, "address deleted")
	}
}
