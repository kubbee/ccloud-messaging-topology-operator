package services

import (
	"os/exec"
)

//
func setEnvironment(environmentId string) (bool, error) {
	cmd := exec.Command("/bin/confluent", "environment", "use", environmentId)

	if err := cmd.Run(); err != nil {
		return false, err
	}

	return true, nil
}
