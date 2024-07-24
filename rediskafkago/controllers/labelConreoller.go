package controllers

import (
	"github.com/labstack/echo/v4"
	"github.com/redis/go-redis/v9"
	"net/http"
	"rediskafkago/models"
)

// получить метки со всеми сязананами адресами
func GetLabel(db *models.MetadataDB[models.Label]) echo.HandlerFunc {
	return func(c echo.Context) error {
		label := c.Param("label")
		data, err := db.Get(label)
		if err != nil {
			if err == redis.Nil {
				return c.JSON(http.StatusNotFound, "The requested  resource does not exit.")
			} else {
				return c.JSON(http.StatusInternalServerError, err.Error())
			}
		}

		return c.JSON(http.StatusOK, data)
	}
}

func GetLabels(db *models.MetadataDB[models.Label]) echo.HandlerFunc {
	return func(c echo.Context) error {
		values, err := db.GetAll(client)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, err.Error())
		}

		return c.JSON(http.StatusOK, values)
	}
}
