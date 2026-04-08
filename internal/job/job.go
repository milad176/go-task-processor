package job

type Job struct {
	ID      string
	Type    string
	Payload map[string]interface{}
	Status  string
}
