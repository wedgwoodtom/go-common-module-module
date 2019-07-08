// +build integration

package awsClients

import (
	_ "fmt"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sqs"
	"math/rand"
	"testing"
	"time"
)

// These are integration tests and not unit tests. Inorder for these to run locally make sure your aws creds file is set with access key and secret key(profile : test-account) from the service user in AWS Test Account
// secret and access key can be found in vault at the path urcs/lab/svc-cts-tvp-entitlement-reimagine-test in end_user namespace

const (
	sqs_queue = "urcs"
)

func createSQSMessageSession() SQSMessageQueue {

	// Create SQSMessageQueue instance
	sess := session.Must(session.NewSession())
	sqsMessageQueue := NewQueue(sqs.New(sess), sqs_queue)

	return sqsMessageQueue

}

func TestSendAndReceiveMessage(t *testing.T) {

	sqsMessageQueue := createSQSMessageSession()

	messageBody := randomString(10)

	//Send Message
	if _, err := sqsMessageQueue.SendMessage(messageBody); err != nil {
		t.Errorf("%d, unexpected error", err)
	}

	//receive message
	messages, err := sqsMessageQueue.ReceiveMessage(0, UseAllAttribute())
	if err != nil {
		t.Errorf("%d, unexpected error", err)
	}

	t.Log(messages)

	for _, m := range messages {

		if *m.Body == messageBody {
			sqsMessageQueue.DeleteMessage(m.ReceiptHandle)
		}
	}

}

func TestSendAndReceiveMessageWithAttributes(t *testing.T) {
	sqsMessageQueue := createSQSMessageSession()

	// Set MessageAttributes
	attrs := map[string]interface{}{
		"source": "abc",
	}

	messageBody := randomString(10)

	//Send Message
	if _, err := sqsMessageQueue.SendMessage(messageBody, MessageAttributes(attrs)); err != nil {
		t.Errorf("%d, unexpected error", err)
	}

	//receive message
	messages, err := sqsMessageQueue.ReceiveMessage(0, UseMessageAttributes("source"), MaxNumberOfMessages(10))
	if err != nil {
		t.Errorf("%d, unexpected error", err)
	}

	t.Log(messages)

	for _, m := range messages {
		if *m.Body == messageBody {
			if _, exist := m.MessageAttributes["source"]; exist {
			} else {
				t.Fail()
			}
			sqsMessageQueue.DeleteMessage(m.ReceiptHandle)
		}
	}
}

func randomString(length int) string {
	charset := "abcdefghijklmnopqrstuvwxyz" +
		"ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	return stringWithCharset(length, charset)
}

func stringWithCharset(length int, charset string) string {
	var seededRand *rand.Rand = rand.New(
		rand.NewSource(time.Now().UnixNano()))

	b := make([]byte, length)
	for i := range b {
		b[i] = charset[seededRand.Intn(len(charset))]
	}
	return string(b)
}
