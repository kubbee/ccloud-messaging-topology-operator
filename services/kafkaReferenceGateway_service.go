package services

import (
	"errors"

	"github.com/go-logr/logr"
)

func BuildKafkaReference(environmentId string, clusterName string, logger *logr.Logger) (*string, error) {
	logger.Info("start::BuildKafkaReference")

	if success, err := setEnvironment(environmentId, logger); err != nil {
		logger.Error(err, "error to select the environment")
	} else {
		if *success {
			return getKafkaCluster(clusterName, logger)
		}
	}

	var blank string = ""
	err := errors.New("was not possible set the environment")

	return &blank, err
}
