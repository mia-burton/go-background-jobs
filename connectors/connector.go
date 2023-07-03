package connectors

type Connector interface {
	Add(queueName string, id string, data []byte) error
	GetQueueLength(queueName string) (int64, error)
	GetFirstElement(queueName string) (string, []byte, error)
	RemoveById(queueName string, id string) error
	RemoveFirstElement(queueName string) ([]byte, error)
	Trim(queueName string, start int64, end int64) error
	NewProgressiveId(queueName string) (string, error)
}
