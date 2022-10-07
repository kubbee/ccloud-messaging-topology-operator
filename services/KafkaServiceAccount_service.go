package services

import (
	"bytes"
	"encoding/json"
	"os/exec"

	"github.com/go-logr/logr"
	util "github.com/kubbee/ccloud-messaging-topology-operator/internal"
)

// createServiceAccount this func create a service account into confluent cloud envrionment
func createServiceAccount(sa, description string, logger *logr.Logger) (*util.ServiceAccount, error) {
	//confluent iam service-account create CadatralServiceAccount --description "This is a text"
	logger.Info("start::createServiceAccoun")

	cmd := exec.Command("/bin/confluent", "iam", "service-account", "create", sa, "--description", description, "--output", "json")

	cmdOutput := &bytes.Buffer{}
	cmd.Stdout = cmdOutput

	if err := cmd.Run(); err != nil {
		logger.Error(err, "error to create service account "+sa)
		return &util.ServiceAccount{}, err
	} else {

		output := cmdOutput.Bytes()
		message, _ := getOutput(output)

		// print message on the log
		logger.Info(message)

		serviceAccount := util.ServiceAccount{}
		json.Unmarshal([]byte(message), &serviceAccount)

		logger.Info("ServiceAccount was created")

		return &serviceAccount, nil
	}

}
