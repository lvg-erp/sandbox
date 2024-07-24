package utils

import (
	"fmt"
	"rediskafkago/models"
)

func HasDuplication(target string, arr []string) bool {
	for _, item := range arr {
		if item == target {
			return true
		}
	}

	return false
}

// utils
// хранилище записей по адресам
func StoreAddressLabels(address string, labels []string, db *models.MetadataDB[models.Label]) error {
	fmt.Println("db related labels")
	for _, label := range labels {
		//проверим дубли
		exists, err := db.KeyExists(label)
		if err != nil {
			return err
		}

		// проверяем, существует ли метка или адрес уже существует в адресе метки
		// метка уже существует, добавим адрес к метке, если адрес не существует в адресе адреса метки

		if exists {
			value, err := db.Get(label)
			if err != nil {
				return err
			}
			arr := value.Address
			exist := HasDuplication(address, arr)
			if exist { // адрес уже есть
				return fmt.Errorf("label %s already exists", label)
			} else {
				value.Address = append(value.Address, address)
				//добавим запись в базу
				if err := db.Update(label, value); err != nil {
					return err
				}
			}
		}

		// метка не существует, сохраняем метку с адресом в БД
		var data models.Label
		data.Address = append(data.Address, address)
		data.Label = label
		if err := db.Update(label, data); err != nil {
			return err
		}
	}

	return nil
}
