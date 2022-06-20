package services

import (
	"bytes"
	"encoding/json"
	"errors"
	"os/exec"

	"github.com/go-logr/logr"
	util "github.com/kubbee/ccloud-messaging-topology-operator/internal"
)

//
func getKafkaCluster(kafkaClusterName string, logger *logr.Logger) (*string, error) {
	logger.Info("start::getKafkaCluster")

	var clusterId string = ""

	cmd := exec.Command("/bin/confluent", "kafka", "cluster", "list", "--output", "json")

	cmdOutput := &bytes.Buffer{}
	cmd.Stdout = cmdOutput

	if err := cmd.Run(); err != nil {
		logger.Error(err, "error to select kafka cluster")
		return &clusterId, err
	} else {

		output := cmdOutput.Bytes()
		message, _ := getOutput(output)

		clusters := []util.ClusterKafka{}

		json.Unmarshal([]byte(message), &clusters)

		for i := 0; i < len(clusters); i++ {
			if kafkaClusterName == clusters[i].Name {
				clusterId = clusters[i].Id
				break
			}
		}

		if clusterId == "" {
			// create an error
			e := errors.New("kafka cluster informed not exists")

			logger.Error(e, "the kafka cluster name was informed on CRD not exists")
			return &clusterId, e
		}

		return &clusterId, nil
	}
}
