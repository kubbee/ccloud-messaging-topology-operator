package services

import (
	"bytes"
	"os/exec"
	"regexp"
)

/*
 * This function get the output console and converts to string
 */
func getOutput(outs []byte) (string, bool) {
	if len(outs) > 0 {
		return string(outs), true
	}
	return "", false
}

/*
 * This function verify the existence of a topic create with the name
 */
func existsTopicName(clusterId string, topicName string) (bool, error) {

	cmd := exec.Command("/bin/confluent", "kafka", "topic", "list", "--cluster", clusterId)

	cmdOutput := &bytes.Buffer{}
	cmd.Stdout = cmdOutput

	if err := cmd.Run(); err != nil {
		//c.Log.Error(err, "Falied to verify the existence of topic")
		return false, err
	} else {

		output := cmdOutput.Bytes()
		message, _ := getOutput(output)

		return regexp.MatchString("\\b"+topicName+"\\b", message)
	}
}

/*
 * This function will generates the name of the kafka topic, if in the tenant name you pass something
 * different of blank and default, the name will be use the tenant do compose the toppic name;
 */
func topicNameGenerator(tenant string, namespace string, topicName string) string {
	separator := "-"
	x := string("")
	if tenant != x || tenant != string("default") {
		return tenant + separator + namespace + separator + topicName
	}
	return namespace + separator + topicName
}
