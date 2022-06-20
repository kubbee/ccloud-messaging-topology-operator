package services

import (
	"github.com/go-logr/logr"
	util "github.com/kubbee/ccloud-messaging-topology-operator/internal"
)

var message string = ""

func BuildTopic(topic *util.NewTopic, environmentId string, clusterId string, logger *logr.Logger) (*string, error) {
	logger.Info("start::BuildTopic")

	_, err := setEnvironment(environmentId, logger)

	if err != nil {
		logger.Error(err, "erro to selecet the environment")
		return &message, err
	}

	return createNewTopic(topic, clusterId, logger)
}
