package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/IBM/sarama"
	"log"
	"net/http"
	"rediskafkago/config"
	"rediskafkago/models"
	"rediskafkago/utils"
)

type ConsumerGroupHandler struct {
	addressDB     *models.MetadataDB[models.Address]
	labelDB       *models.MetadataDB[models.Label]
	transactionDB *models.MetadataDB[models.Transaction]
}

func (*ConsumerGroupHandler) Setup(sarama.ConsumerGroupSession) error   { return nil } //требуется для удовлетворения интерфейса sarama.ConsumerGroupHandler
func (*ConsumerGroupHandler) Cleanup(sarama.ConsumerGroupSession) error { return nil }

func (handler *ConsumerGroupHandler) ConsumeClaim(sess sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for msg := range claim.Messages() {
		topic := msg.Topic
		key := string(msg.Key)
		log.Printf("Consuming topic: %s", topic)

		switch topic {
		case "address":
			var data models.Address
			if err := json.Unmarshal(msg.Value, &data); err != nil {
				log.Printf("failsed unmarshal address: %v", err)
			}

			if err := handler.addressDB.Update(key, data); err != nil {
				log.Printf("failsed to update address: %v", err)
			}

			if err := utils.StoreAddressLabels(key, data.Labels, handler.labelDB); err != nil {
				log.Printf("failed to store labels: %v", err)
			}

			sess.MarkMessage(msg, "") // mark the msg as processed
		case "label":
			var data models.Label
			if err := json.Unmarshal(msg.Value, &data); err != nil {
				log.Printf("failed to unmarshal address: %v", err)
			}

			if err := handler.labelDB.Update(key, data); err != nil {
				log.Printf("failed to update db: %v", err)
			}
			sess.MarkMessage(msg, "") // mark the msg as processed
		case "transaction":
			var data models.Transaction
			if err := json.Unmarshal(msg.Value, &data); err != nil {
				log.Printf("failed to unmarshall transaction: %v")
			}
			if err := handler.transactionDB.Update(key, data); err != nil {
				log.Printf("failed to updae db: %v", err)
			}
			sess.MarkMessage(msg, "")
		}
	}

	return nil
}

func initConsumerGroup() (sarama.ConsumerGroup, error) {

	saramaConfig := sarama.NewConfig()
	consumerGroup, err := sarama.NewConsumerGroup([]string{config.KafkaServerAddr}, config.ConsumerGroup, saramaConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize consumer group: %w", err)
	}

	return consumerGroup, nil

}

func setUpConsumerGroup(ctx context.Context,
	addressDB *models.MetadataDB[models.Address],
	labelDB *models.MetadataDB[models.Label],
	transactionDB *models.MetadataDB[models.Transaction]) {

	consumerGroup, err := initConsumerGroup()
	if err != nil {
		log.Printf("consumer initialization error: %v", err)
	}

	defer consumerGroup.Close()

	consumer := &ConsumerGroupHandler{
		addressDB:     addressDB,
		labelDB:       labelDB,
		transactionDB: transactionDB,
	}

	for { // цикл бесконечно. обрабатывать сообщения из темы и обрабатывать любые возникающие ошибки
		if err = consumerGroup.Consume(ctx, config.Topics[:], consumer); err != nil {
			log.Printf("error from consumer: %v", err)
		}
		if ctx.Err() != nil {
			return
		}

	}

}

func main() {
	ctx, cancel := context.WithCancel(context.Background())

	go setUpConsumerGroup(ctx, models.AddressDB, models.LabelDB, models.TransactionDB)
	defer cancel()

	fmt.Printf("Kafka CONSUMER (Group: %s) 👥📥 "+
		"started at http://localhost%s\n", config.ConsumerGroup, config.ConsumerPort)

	// Use a blocking operation to serve HTTP requests, which keeps the program running
	if err := http.ListenAndServe(":8081", nil); err != nil {
		log.Fatal(err)
	}
}
