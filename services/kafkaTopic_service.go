package services

import (
	"os/exec"

	"github.com/go-logr/logr"
	util "github.com/kubbee/ccloud-messaging-topology-operator/internal"
)

func createNewTopic(topic *util.NewTopic, clusterId string, logger *logr.Logger) (*string, error) {
	logger.Info("Start::CreateNewTopic")

	// generates the topic name
	topicName := topicNameGenerator(topic.Tenant, topic.Namespace, topic.Topic)

	if isOk, err := existsTopicName(clusterId, topicName); err != nil {
		logger.Error(err, "error to verify if topic already exists")
		return &topicName, err
	} else {
		if !isOk {
			cmd := exec.Command("/bin/confluent", "kafka", "topic", "create", topicName, "--partitions", topic.Partitions, "--cluster", clusterId)

			if err := cmd.Run(); err != nil {
				logger.Error(err, "Confluent::NewTopic:: Error to run the command to create kafka topic")
				return &topicName, err
			} else {
				return &topicName, nil
			}
		} else {
			return &topicName, nil
		}
	}
}
