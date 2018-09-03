package db

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"reflect"

	"github.com/renevo/mcstats/pkg/logging"
	bolt "go.etcd.io/bbolt"
)

type boltDB struct {
	inner *bolt.DB
}

func (b *boltDB) Close() error {
	return b.inner.Close()
}

func (b *boltDB) Get(partition, name string, v interface{}) error {
	logging.Debug("GET %s/%s", partition, name)

	// validation check, this is a programmer error, so we panic...
	if reflect.ValueOf(v).Kind() != reflect.Ptr {
		panic("value must be pointer")
	}

	return b.inner.View(func(t *bolt.Tx) error {
		// TODO: support path in partition for sub-buckets...
		bucket := t.Bucket([]byte(partition))
		if bucket == nil {
			return fmt.Errorf("failed to load partition %s: not found", partition)
		}

		data := bucket.Get([]byte(name))
		if len(data) == 0 {
			return nil
		}

		buff := bytes.NewBuffer(data)
		if err := gob.NewDecoder(buff).Decode(v); err != nil {
			return fmt.Errorf("failed to unmarshal value on partition %s with name %s: %v", partition, name, err)
		}

		return nil
	})
}

func (b *boltDB) Put(partition, name string, v interface{}) error {
	logging.Debug("PUT %s/%s - %+v", partition, name, v)

	return b.inner.Update(func(t *bolt.Tx) error {
		// TODO: support path in partition for sub-buckets...
		bucket, err := t.CreateBucketIfNotExists([]byte(partition))
		if err != nil {
			return fmt.Errorf("failed to create partition %s: %v", partition, err)
		}

		// TODO: for funsies, because i want to, make this a sync.Pool
		var buff bytes.Buffer

		if err := gob.NewEncoder(&buff).Encode(v); err != nil {
			return fmt.Errorf("failed to marshal value %+v on partition %s with name %s: %v", v, partition, name, err)
		}

		return bucket.Put([]byte(name), buff.Bytes())
	})
}
