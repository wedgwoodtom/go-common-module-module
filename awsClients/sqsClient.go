package awsClients

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/aws/aws-sdk-go/service/sqs/sqsiface"
	"log"
	"strconv"
)

// The DataType is a type of data used in Attributes and Message Attributes.
const (
	DataTypeString = "String"
	DataTypeNumber = "Number"
	DataTypeBinary = "Binary"
)

// SQSMessageQueue provides the ability to handle Client messages.
type SQSMessageQueue struct {
	Client   sqsiface.SQSAPI
	QueueUrl *string
}

// A BatchMessage represents each request to send a message.
// SendMessagesInput are used to change parameters for the message.
type BatchMessage struct {
	Body              string
	SendMessagesInput []SendMessageInput
}

// NewQueue initializes SQSMessageQueue with queue url and client session .
func NewQueue(svc sqsiface.SQSAPI, name string) SQSMessageQueue {
	u, err := GetQueueURL(svc, name)
	if err != nil {
		log.Fatal(err)

	}

	return SQSMessageQueue{
		Client:   svc,
		QueueUrl: u,
	}
}

// GetQueueURL returns a QueueUrl for the given queue name. If an error occurs that error will be returned.
func GetQueueURL(s sqsiface.SQSAPI, name string) (*string, error) {
	req := &sqs.GetQueueUrlInput{
		QueueName: aws.String(name),
	}

	resp, err := s.GetQueueUrl(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get queue url, %v", err)
	}
	return resp.QueueUrl, nil
}

// ReceiveMessage returns the messages from Client if any. If an error occurs that error will be returned.
func (q *SQSMessageQueue) ReceiveMessage(waitTimeout int64, opts ...ReceiveMessageInput) ([]*sqs.Message, error) {
	req := &sqs.ReceiveMessageInput{
		QueueUrl: q.QueueUrl,
	}
	if waitTimeout > 0 {
		req.WaitTimeSeconds = aws.Int64(waitTimeout)
	}

	for _, f := range opts {
		f(req)
	}
	resp, err := q.Client.ReceiveMessage(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get messages, %v", err)
	}

	return resp.Messages, nil
}

// SendMessageBatch sends messages to SQS queue.
func (q *SQSMessageQueue) SendMessageBatch(messages ...BatchMessage) (*sqs.SendMessageBatchOutput, error) {
	entries := BuildBatchRequestEntry(messages...)

	req := &sqs.SendMessageBatchInput{
		Entries:  entries,
		QueueUrl: q.QueueUrl,
	}

	resp, err := q.Client.SendMessageBatch(req)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

// BuildBatchRequestEntry builds batch entries.
func BuildBatchRequestEntry(messages ...BatchMessage) []*sqs.SendMessageBatchRequestEntry {
	entries := make([]*sqs.SendMessageBatchRequestEntry, len(messages))
	for i, bm := range messages {
		req := &sqs.SendMessageInput{}
		for _, sendMessageInput := range bm.SendMessagesInput {
			sendMessageInput(req)
		}

		id := aws.String(fmt.Sprintf("msg-%d", i))
		entries[i] = &sqs.SendMessageBatchRequestEntry{
			DelaySeconds:      req.DelaySeconds,
			MessageAttributes: req.MessageAttributes,
			MessageBody:       aws.String(bm.Body),
			Id:                id,
		}
	}

	return entries
}

// DeleteMessage deletes a message from Client queue. If an error occurs that error will be returned.
func (q *SQSMessageQueue) DeleteMessage(receiptHandle *string) error {
	_, err := q.Client.DeleteMessage(&sqs.DeleteMessageInput{
		QueueUrl:      q.QueueUrl,
		ReceiptHandle: receiptHandle,
	})
	return err
}

// SendMessage sends a message to Client queue. If an error occurs that error will be returned.
func (q *SQSMessageQueue) SendMessage(body string, opts ...SendMessageInput) (*sqs.SendMessageOutput, error) {
	req := &sqs.SendMessageInput{
		MessageBody: aws.String(body),
		QueueUrl:    q.QueueUrl,
	}

	for _, sendMessageInput := range opts {
		sendMessageInput(req)
	}

	resp, err := q.Client.SendMessage(req)

	if err != nil {
		return nil, fmt.Errorf("failed to send messages, %v", err)
	}

	return resp, nil
}

// The SendMessageInput type sets a parameter in sqs.SendMessageInput.
type SendMessageInput func(req *sqs.SendMessageInput)

// DelaySeconds returns a SendMessageInput that sets DelaySeconds to delay in seconds.
func DelaySeconds(delay int64) SendMessageInput {
	return func(req *sqs.SendMessageInput) {
		req.DelaySeconds = aws.Int64(delay)
	}
}

// MessageGroupId returns a SendMessageInput that sets MessageGroupId to MessageGroupId.
func MessageGroupId(messageGroupId string) SendMessageInput {
	return func(req *sqs.SendMessageInput) {
		//only works for fifo queue
		req.MessageGroupId = aws.String(messageGroupId)
	}
}

// MessageDeduplicationId returns a SendMessageInput that sets MessageDeduplicationId to MessageDeduplicationId.
func MessageDeduplicationId(messageDeduplicationId string) SendMessageInput {
	return func(req *sqs.SendMessageInput) {
		//only works for fifo queue
		req.MessageDeduplicationId = aws.String(messageDeduplicationId)
	}
}

// MessageAttributes returns a SendMessageInput that sets MessageAttributes to attrs.
// A string value in attrs sets to DataTypeString.
// A []byte value in attrs sets to DataTypeBinary.
// A int and int64 value in attrs sets to DataTypeNumber. Other types cause panicking.
func MessageAttributes(attrs map[string]interface{}) SendMessageInput {
	return func(req *sqs.SendMessageInput) {
		if len(attrs) == 0 {
			return
		}

		ret := make(map[string]*sqs.MessageAttributeValue)
		for n, v := range attrs {
			ret[n] = MessageAttributeValue(v)
		}
		req.MessageAttributes = ret
	}
}

// MessageAttributeValue returns a appropriate sqs.MessageAttributeValue by type assersion of v.
// Types except string, []byte, int64 and int cause panicking.
func MessageAttributeValue(v interface{}) *sqs.MessageAttributeValue {
	switch vv := v.(type) {
	case string:
		return &sqs.MessageAttributeValue{
			DataType:    aws.String(DataTypeString),
			StringValue: aws.String(vv),
		}
	case []byte:
		return &sqs.MessageAttributeValue{
			DataType:    aws.String(DataTypeBinary),
			BinaryValue: vv,
		}
	case int64:
		return &sqs.MessageAttributeValue{
			DataType:    aws.String(DataTypeNumber),
			StringValue: aws.String(strconv.FormatInt(vv, 10)),
		}
	case int:
		return &sqs.MessageAttributeValue{
			DataType:    aws.String(DataTypeNumber),
			StringValue: aws.String(strconv.FormatInt(int64(vv), 10)),
		}
	default:
		panic("sqs: unsupported type")
	}
}

// The ReceiveMessageInput type sets a parameter in sqs.ReceiveMessageInput.
type ReceiveMessageInput func(req *sqs.ReceiveMessageInput)

// MaxNumberOfMessages returns a ReceiveMessageInput that changes a max number of messages to receive to n.
func MaxNumberOfMessages(n int64) ReceiveMessageInput {
	return func(req *sqs.ReceiveMessageInput) {
		req.MaxNumberOfMessages = aws.Int64(n)
	}
}

// UseAllAttribute returns a ReceiveMessageInput that changes a parameter to receive all messages regardless of attributes.
func UseAllAttribute() ReceiveMessageInput {
	return UseAttributes("All")
}

// UseAttributes returns a ReceiveMessageInput that sets AttributeNames to attr.
func UseAttributes(attr ...string) ReceiveMessageInput {
	return func(req *sqs.ReceiveMessageInput) {
		req.AttributeNames = aws.StringSlice(attr)
	}
}

// UseMessageAttributes returns a ReceiveMessageInput that sets MessageAttributeNames to attr.
func UseMessageAttributes(attr ...string) ReceiveMessageInput {
	return func(req *sqs.ReceiveMessageInput) {
		req.MessageAttributeNames = aws.StringSlice(attr)
	}
}
