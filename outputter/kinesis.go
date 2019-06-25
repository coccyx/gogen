package outputter

import (
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/kinesis"
	config "github.com/coccyx/gogen/internal"
	"math/rand"
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
		k.endpoint = item.S.Output.Endpoints[rand.Intn(len(item.S.Output.Endpoints))]
	}

	for _, e := range item.Events {
		partkey := e["host"]

		evt := e["_raw"]

		if evt[len(evt) - 1] != '\n' {
			evt = evt + "\n"
		}
		rec := kinesis.PutRecordsRequestEntry{
			PartitionKey: &partkey,
			Data: []byte(evt),
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
	if len(k.buf) == 0 {
		return nil
	}
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

	results, e := k.client.PutRecords(&kinesisRequest)

	if e == nil {
		print(*results.FailedRecordCount)
		print("\n")
	}

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