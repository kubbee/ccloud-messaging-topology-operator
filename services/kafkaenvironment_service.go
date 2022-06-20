package services

import (
	"os/exec"

	"github.com/go-logr/logr"
)

//
func setEnvironment(environmentId string, logger *logr.Logger) (*bool, error) {
	logger.Info("start::setEnvironment")
	var success bool = false
	cmd := exec.Command("/bin/confluent", "environment", "use", environmentId)

	if err := cmd.Run(); err != nil {
		logger.Error(err, "Error select the environment")
		return &success, err
	}

	return &success, nil
}
