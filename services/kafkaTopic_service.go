package services

import (
	"os/exec"
)

func createNewTopic(tenant string, topic string, namespace string, partitions string, clusterId string) (*string, error) {
	// generates the topic name
	name := topicNameGenerator(tenant, namespace, topic)

	if ok, err := existsTopicName(clusterId, name); err != nil {
		return &name, err
	} else {
		if ok {
			return &name, nil
		} else {
			cmd := exec.Command("/bin/confluent", "kafka", "topic", "create", name, "--partitions", partitions, "--cluster", clusterId)

			if err := cmd.Run(); err != nil {
				return &name, err
			} else {
				return &name, nil
			}
		}
	}
}
