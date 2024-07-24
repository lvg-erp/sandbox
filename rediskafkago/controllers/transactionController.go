package controllers

import (
	"encoding/json"
	"fmt"
	"github.com/IBM/sarama"
	"github.com/labstack/echo/v4"
	"github.com/redis/go-redis/v9"
	"net/http"
	"rediskafkago/models"
)

// GET `transactiones/{hash}`
func GetTransaction(db *models.MetadataDB[models.Transaction]) echo.HandlerFunc {
	return func(c echo.Context) error {
		hash := c.Param("hash")
		data, err := db.Get(hash)
		if err != nil {
			if err == redis.Nil {
				return c.JSON(http.StatusNotFound, "The requested resource does not exist.")
			} else {
				return c.JSON(http.StatusInternalServerError, err.Error())
			}
		}

		return c.JSON(http.StatusOK, data)
	}
}

// // GET `/transactiones`
func GetTransactions(db *models.MetadataDB[models.Transaction]) echo.HandlerFunc {
	return func(c echo.Context) error {
		values, err := db.GetAll(client)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, err.Error())
		}
		return c.JSON(http.StatusOK, values)
	}
}

// POST `/transactiones/{hash}`
// создаем транзакцию
func CreateTransaction(producer sarama.SyncProducer, db *models.MetadataDB[models.Transaction]) echo.HandlerFunc {
	return func(c echo.Context) error {
		hash := c.Param("hash")
		exists, err := db.KeyExists(hash)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, err.Error())
		}
		//Проверим возможно уже есть запись с таким ключом
		if exists {
			return c.JSON(http.StatusBadRequest, "The resource to create already exist.")
		}

		var data models.Transaction
		//получим параметры из заголовка
		if err := c.Bind(&data); err != nil {
			return c.JSON(http.StatusBadRequest, err.Error())
		}
		//валидируем
		if validationErr := validate.Struct(&data); validationErr != nil {
			return c.JSON(http.StatusBadRequest, validationErr.Error())
		}
		//проверим
		if hash != data.Hash {
			return c.JSON(http.StatusBadRequest, "transaction hash not match")
		}

		newTx := models.Transaction{
			Hash:    data.Hash,
			Chainid: data.Chainid,
			From:    data.From,
			To:      data.To,
			Status:  data.Status,
		}
		//отправляем в kafka
		newTxJSON, err := json.Marshal(newTx)
		if err != nil {
			return fmt.Errorf("failed to marshall: %w", err)
		}

		msg := &sarama.ProducerMessage{
			Topic: "transaction",
			Key:   sarama.StringEncoder(hash),
			Value: sarama.StringEncoder(newTxJSON),
		}

		if _, _, err := producer.SendMessage(msg); err != nil {
			return c.JSON(http.StatusInternalServerError, err.Error())
		}

		return c.JSON(http.StatusOK, "transaction create")
	}
}

// PUT `/transactiones/{hash}`
// update a transaction

func UpdateTransaction(producer sarama.SyncProducer, db *models.MetadataDB[models.Transaction]) echo.HandlerFunc {
	return func(c echo.Context) error {
		hash := c.Param("hash")
		exists, err := db.KeyExists(hash)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, err.Error())
		}
		if !exists {
			return c.JSON(http.StatusBadRequest, "The resource to upgrade not exist")
		}

		var data models.Transaction
		if err := c.Bind(&data); err != nil {
			return c.JSON(http.StatusBadRequest, err.Error())
		}
		//валидация
		if validationErr := validate.Struct(&data); validationErr != nil {
			return c.JSON(http.StatusBadRequest, validationErr.Error())
		}
		newTx := models.Transaction{
			Hash:    data.Hash,
			Chainid: data.Chainid,
			From:    data.From,
			To:      data.To,
			Status:  data.Status,
		}

		//отправка в очередь kafka
		newTxJSON, err := json.Marshal(newTx)
		if err != nil {
			return fmt.Errorf("failed to marshal transaction: %w", err)
		}

		msg := &sarama.ProducerMessage{
			Topic: "transaction",
			Key:   sarama.StringEncoder(hash),
			Value: sarama.StringEncoder(newTxJSON),
		}

		if _, _, err := producer.SendMessage(msg); err != nil {
			return c.JSON(http.StatusInternalServerError, err.Error())
		}

		return c.JSON(http.StatusOK, "transaction update")
	}
}

// DELETE `/transactiones/{hash}`
// delete a transaction

func DeleteTransaction(producer sarama.SyncProducer, db *models.MetadataDB[models.Transaction]) echo.HandlerFunc {
	return func(c echo.Context) error {
		hash := c.Param("hash")
		keyDeleted, err := db.Delete(hash)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, err.Error())
		}
		if keyDeleted == 0 {
			return c.JSON(http.StatusBadRequest, "The resource to delete does not exist.")
		}

		return c.JSON(http.StatusOK, "transaction deleted")
	}
}
