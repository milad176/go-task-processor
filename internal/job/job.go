package job

import "fmt"

type Job struct {
	ID         string
	Type       string
	Payload    map[string]interface{}
	Status     string
	Retries    int
	MaxRetries int
}

func (j *Job) Validate() error {
	if j.ID != "" {
		return fmt.Errorf("ID should not be provided")
	}

	if j.Type == "" {
		return fmt.Errorf("type is required")
	}

	switch j.Type {
	case "print":
		if j.Payload["message"] == "" {
			return fmt.Errorf("message is required for print job")
		}

	case "sleep":
		if j.Payload["duration"] == "" {
			return fmt.Errorf("duration is required for sleep job")
		}

	case "fail":
		if j.Payload["message"] == "" {
			return fmt.Errorf("message is required for fail job")
		}

	default:
		return fmt.Errorf("invalid job type: %s", j.Type)
	}

	return nil
}
