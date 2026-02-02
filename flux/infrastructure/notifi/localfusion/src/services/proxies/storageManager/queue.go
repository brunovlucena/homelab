package storageManager

type Queue struct {
	values []string
}

func NewQueue() *Queue {
	return &Queue{
		values: make([]string, 0),
	}
}

func (q *Queue) Enqueue(value string) {
	q.values = append(q.values, value)
}

func (q *Queue) Dequeue() *string {
	if len(q.values) == 0 {
		return nil
	}

	value := q.values[0]
	q.values = q.values[1:]
	return &value
}

func (q *Queue) Peek() *string {
	if len(q.values) == 0 {
		return nil
	}

	value := q.values[0]
	return &value
}
