package services

import (
	"os/exec"

	"github.com/go-logr/logr"
)

func createNewTopic(tenant string, topic string, namespace string, partitions string, clusterId string, logger *logr.Logger) (*string, error) {
	logger.Info("Start::CreateNewTopic")

	// generates the topic name
	name := topicNameGenerator(tenant, namespace, topic)

	if isOk, err := existsTopicName(clusterId, name); err != nil {
		logger.Error(err, "error to verify if topic already exists")
		return &name, err
	} else {
		if !isOk {
			cmd := exec.Command("/bin/confluent", "kafka", "topic", "create", name, "--partitions", partitions, "--cluster", clusterId)

			if err := cmd.Run(); err != nil {
				logger.Error(err, "Confluent::NewTopic:: Error to run the command to create kafka topic")
				return &name, err
			} else {
				logger.Info("Topic was created")
				return &name, nil
			}
		} else {
			logger.Info("Topic was retrieved")
			return &name, nil
		}
	}
}
