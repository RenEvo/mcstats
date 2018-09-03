package mcstats

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// StatsGroups represents the stats collection
type StatsGroups map[string]Stats

// Stats represents a collection of stats
type Stats map[string]int

// StatsFile represents the stats file
type StatsFile struct {
	ID          string      `json:"id"`
	Name        string      `json:"name"`
	Stats       StatsGroups `json:"stats"`
	Timestamp   time.Time   `json:"timestamp"`
	DataVersion int
}

// Load will parse all player stats files in the given directory
func Load(statsPath string) ([]StatsFile, error) {
	absPath, err := filepath.Abs(statsPath)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve absolute path for %s: %v", statsPath, err)
	}

	statsFiles := []StatsFile{}

	if err := filepath.Walk(absPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		if !strings.EqualFold(filepath.Ext(info.Name()), ".json") {
			return nil
		}

		statFile, err := New(path)
		if err != nil {
			return fmt.Errorf("failed to create stats for %s: %v", path, err)
		}

		statsFiles = append(statsFiles, *statFile)

		return nil
	}); err != nil {
		return nil, fmt.Errorf("unable to walk stats directory %s: %v", absPath, err)
	}

	return statsFiles, nil
}

// New takes the supplied mc stats file, looks up the user, and parses the data
func New(file string) (*StatsFile, error) {
	info, err := os.Stat(file)
	if err != nil {
		return nil, fmt.Errorf("failed to stat file %s: %v", file, err)
	}

	uid := info.Name()[:len(info.Name())-len(".json")]

	stats := &StatsFile{}
	data, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, fmt.Errorf("failed to read file %s: %v", file, err)
	}

	if err := json.Unmarshal(data, stats); err != nil {
		return nil, fmt.Errorf("failed to deserialize stats %s: %v", file, err)
	}

	stats.Timestamp = info.ModTime()
	stats.ID = uid

	name, err := GetName(uid, info.ModTime())
	if err != nil {
		// Should we fail on this?
		return nil, fmt.Errorf("failed to resolve name for uid %s: %v", uid, err)
	}

	stats.Name = name

	return stats, nil
}

// ToLineFormat returns an Influx Line Format
func (s StatsFile) ToLineFormat() string {
	metrics := ""
	tsn := s.Timestamp.UnixNano()

	for group, gstats := range s.Stats {
		for stat, value := range gstats {
			metrics = fmt.Sprintf("%s%s,uid=%s,user=%s,name=%s,version=%d total=%d %d\n", metrics, group, s.ID, s.Name, stat, s.DataVersion, value, tsn)
		}
	}

	return metrics
}

func (s StatsFile) String() string {
	return fmt.Sprintf("ID: %s; User: %s; Version: %d; Last Updated: %v", s.ID, s.Name, s.DataVersion, s.Timestamp)
}
