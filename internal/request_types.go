package internal

type ACLRequest struct {
	ServiceAccount string
	Authorize      string
	Operation      string
	Topic          string
	ClusterId      string
	EnvironmentId  string
}
