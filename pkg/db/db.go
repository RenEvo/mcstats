package db

import (
	"fmt"
	"io"
	"os"
	"path"
	"time"

	bolt "go.etcd.io/bbolt"
)

const databaseFileMode = 0644

// Database represents the database....
type Database interface {
	io.Closer

	// Put will store the specified value with the name in the partition
	Put(partition, name string, v interface{}) error
	// Get will retrieve the stored value with the name in the partition to the specified value
	//
	// The value of v should be passed in as a pointer
	Get(partition, name string, v interface{}) error
}

// Open the specified database from path
func Open(dbPath string) (Database, error) {
	if err := os.MkdirAll(path.Dir(dbPath), databaseFileMode); err != nil {
		return nil, fmt.Errorf("failed to create database directory %s: %v", path.Dir(dbPath), err)
	}

	bdb, err := bolt.Open(dbPath, databaseFileMode, &bolt.Options{Timeout: 1 * time.Minute})
	if err != nil {
		return nil, fmt.Errorf("failed to open boltdb at path %s: %v", dbPath, err)
	}

	return &boltDB{inner: bdb}, nil
}
