package services

import (
	"bytes"
	"encoding/json"
	"errors"
	"os/exec"

	util "github.com/kubbee/ccloud-messaging-topology-operator/internal"
)

//
func getKafkaCluster(kafkaClusterName string) (string, error) {
	clusterId := ""

	cmd := exec.Command("/bin/confluent", "kafka", "cluster", "list", "--output", "json")

	cmdOutput := &bytes.Buffer{}
	cmd.Stdout = cmdOutput

	if err := cmd.Run(); err != nil {
		return clusterId, err
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
			return clusterId, errors.New("kafka cluster informed not exists")
		}

		return clusterId, nil
	}
}
