package influx

import (
	"fmt"

	retryablehttp "github.com/hashicorp/go-retryablehttp"
)

// Database represents a remote influx database
type Database struct {
	server string
	db     string
}

// CreateDatabase will attemptto create the specified database with a default retention of 7d
func CreateDatabase(server, db string) (*Database, error) {
	resp, err := retryablehttp.Get(server + "/query?q=CREATE%20DATABASE%20" + db + "%20WITH%20DURATION%207d%20NAME%20weekly")
	if err != nil {
		return nil, err
	}

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return nil, fmt.Errorf("failed to create database %s: %s", db, resp.Status)
	}

	return &Database{server: server, db: db}, nil
}

// WriteData will attempt to writethe specified influx line format to the specified database
func (db *Database) Write(data string) error {
	resp, err := retryablehttp.Post(db.server+"/write?db="+db.db, "text/plain", []byte(data))
	if err != nil {
		return err
	}

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return fmt.Errorf("failed to write data to %s/%s: %s", db.server, db.db, resp.Status)
	}

	return nil
}
