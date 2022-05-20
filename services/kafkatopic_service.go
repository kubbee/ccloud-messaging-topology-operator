package services

import (
	"errors"

	"github.com/go-logr/logr"
	util "github.com/kubbee/ccloud-messaging-topology-operator/internal"
)

func GetTopic(topic *util.NewTopic, logger *logr.Logger) (*util.TopicReference, error) {

	logger.Info("Creating a New Topic")
	selected := SetEnvironment(topic.ClusterReference.EnvironmentId, logger)

	if selected {
		logger.Info("The environment was selected")
		return CreateNewTopic(topic, logger)
	} else {
		logger.Error(errors.New("error to select the environment"), "Error to select the environment")
		return &util.TopicReference{}, errors.New("error to select the environment")
	}
}
