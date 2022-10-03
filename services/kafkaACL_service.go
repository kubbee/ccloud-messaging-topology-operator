package services

import (
	"bytes"
	"os/exec"

	"github.com/go-logr/logr"
	util "github.com/kubbee/ccloud-messaging-topology-operator/internal"
)

// createACL this functions creates an acl configuration for de ServiceAccount
func createACL(r *util.ACLRequest, logger *logr.Logger) (bool, error) {
	// confluent kafka acl create --allow --service-account sa --operation READ --topic name --cluster iuw90
	logger.Info("start:: func createACL")

	cmd := exec.Command("/bin/confluent", "kafka", "acl", "create", r.Authorize,
		"--service-account "+r.ServiceAccount,
		"--operation "+r.Operation,
		"--topic "+r.Topic,
		"--environment "+r.EnvironmentId,
		"--cluster "+r.ClusterId,
		"--output", "json")

	cmdOutput := &bytes.Buffer{}
	cmd.Stdout = cmdOutput

	if err := cmd.Run(); err != nil {
		logger.Error(err, "error to create service account ", "")
		return false, err
	} else {
		output := cmdOutput.Bytes()
		message, _ := getOutput(output)

		// print message on the log
		logger.Info(message)

		return true, nil
	}
}
