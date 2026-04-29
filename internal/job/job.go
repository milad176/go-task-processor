package job

import "fmt"

type Job struct {
	ID         string                 `json:"ID"`
	Type       string                 `json:"Type"`
	Payload    map[string]interface{} `json:"Payload"`
	Status     string                 `json:"Status"`
	Retries    int                    `json:"-"`
	MaxRetries int                    `json:"-"`
	Priority   int                    `json:"-"`
	ClaimedAt  int64                  `json:"-"`
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

	case "payment":
		if j.Payload["orderId"] == "" {
			return fmt.Errorf("orderId is required for payment job")
		}

	case "report":
		if j.Payload["reportName"] == "" {
			return fmt.Errorf("reportName is required for report job")
		}

	default:
		return fmt.Errorf("invalid job type: %s", j.Type)
	}

	return nil
}

func (j *Job) AssignDefaultPriority() {
	switch j.Type {
	case "payment":
		j.Priority = 3
	case "report":
		j.Priority = 2
	case "print":
		j.Priority = 1
	default:
		j.Priority = 1
	}
}
