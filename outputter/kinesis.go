package outputter

import (
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/kinesis"
	config "github.com/coccyx/gogen/internal"
)

type kinesisout struct {
	buf			[]*kinesis.PutRecordsRequestEntry
	client		*kinesis.Kinesis
	initialized	bool
	closed		bool
	endpoint	string
}

func (k *kinesisout) Send(item *config.OutQueueItem) error {
	if !k.initialized {
		k.buf = []*kinesis.PutRecordsRequestEntry{}
		sess, err := session.NewSession()
		if err != nil {
			return err
		}
		k.client = kinesis.New(sess)
	}

	for _, e := range item.Events {
		partkey := e["host"]
		rec := kinesis.PutRecordsRequestEntry{
			PartitionKey: &partkey,
			Data: []byte(e["_raw"]),
		}
		k.buf = append(k.buf, &rec)
	}

	if len(k.buf) >= 500 {
		return k.flush()
	}

	return nil
}

func (k *kinesisout) flush() error {
	var records []*kinesis.PutRecordsRequestEntry
	if len(k.buf) > 500 {
		records = k.buf[:500]
		k.buf = k.buf[500:]
	} else {
		records = k.buf
		k.buf = []*kinesis.PutRecordsRequestEntry{}
	}

	kinesisRequest := kinesis.PutRecordsInput{
		Records: records,
		StreamName: &k.endpoint,
	}

	_, e := k.client.PutRecords(&kinesisRequest)

	return e
}

func (k *kinesisout) Close() error {
	if !k.closed {
		k.closed = true
		err := k.flush()
		if err != nil {
			return err
		}
	}
	return nil
}