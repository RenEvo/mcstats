package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"

	retryablehttp "github.com/hashicorp/go-retryablehttp"
)

var (
	httpClient = retryablehttp.NewClient()
)

// StatsFile is
type StatsFile struct {
	Stats       map[string]map[string]int `json:"stats"`
	DataVersion int
}

func main() {
	statsDirectory := flag.String("dir", "./stats", "Minecraft stats directory")
	influxServer := flag.String("dbserver", "http://localhost:8086", "influx server to connect to")
	influxDB := flag.String("db", "minecraft", "database name")

	flag.Parse()

	if _, err := os.Stat(*statsDirectory); os.IsNotExist(err) {
		panic(fmt.Sprintf("directory %q does not exist", *statsDirectory))
	}

	statsPath, err := filepath.Abs(*statsDirectory)
	if err != nil {
		panic(err)
	}

	if err := createDatabase(*influxServer, *influxDB); err != nil {
		panic(err)
	}

	statsFiles := map[string]string{}
	userList := []string{}
	if err := filepath.Walk(statsPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		if !strings.EqualFold(filepath.Ext(info.Name()), ".json") {
			return nil
		}

		uid := info.Name()[:len(info.Name())-len(".json")]
		userList = append(userList, uid)

		statsFiles[uid] = path

		return nil
	}); err != nil {
		panic(fmt.Sprintf("unable to walk stats directory: %v", err))
	}

	for uid, file := range statsFiles {
		nameHistory := []struct {
			Name    string `json:"name"`
			Changed int64  `json:"changedToAt"`
		}{}

		uidString := strings.ToLower(strings.Replace(uid, "-", "", -1))
		if err := getJSON(fmt.Sprintf("https://api.mojang.com/user/profiles/%s/names", uidString), &nameHistory); err != nil {
			panic(err)
		}

		now := time.Now().UnixNano()

		stats := &StatsFile{}
		data, err := ioutil.ReadFile(file)
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to read file %s: %v\n", file, err)
			continue
		}

		if err := json.Unmarshal(data, stats); err != nil {
			fmt.Fprintf(os.Stderr, "failed to deserialize stats %s: %v\n", file, err)
			continue
		}

		metrics := ""

		// TODO: keep and track increment....
		for group, gstats := range stats.Stats {
			for stat, value := range gstats {
				metrics = fmt.Sprintf("%s%s,uid=%s,user=%s,name=%s,version=%d total=%d %d\n", metrics, group, uid, nameHistory[len(nameHistory)-1].Name, stat, stats.DataVersion, value, now)
			}
		}

		if err := writeData(*influxServer, *influxDB, metrics); err != nil {
			fmt.Fprintf(os.Stderr, "%v\n", err)
			continue
		}
	}
}

func createDatabase(server, name string) error {
	resp, err := retryablehttp.Get(server + "/query?q=CREATE%20DATABASE%20" + name + "%20WITH%20DURATION%207d%20NAME%20weekly")
	if err != nil {
		return err
	}

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return fmt.Errorf("failed to create database %s: %s", name, resp.Status)
	}

	return nil
}

func writeData(server, name, data string) error {
	resp, err := retryablehttp.Post(server+"/write?db="+name, "text/plain", []byte(data))
	if err != nil {
		return err
	}

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return fmt.Errorf("failed to write data to %s: %s", name, resp.Status)
	}

	return nil
}

func getJSON(url string, v interface{}) error {
	req, err := retryablehttp.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}

	req.Header.Set("User-Agent", "mcstats/1.0.0")
	resp, err := httpClient.Do(req)

	if err != nil {
		return err
	}

	defer resp.Body.Close()

	return json.NewDecoder(resp.Body).Decode(v)
}

func postJSON(url string, d interface{}, v interface{}) error {
	reqBody, err := json.Marshal(d)
	if err != nil {
		return err
	}

	req, err := retryablehttp.NewRequest("POST", url, bytes.NewReader(reqBody))
	if err != nil {
		return err
	}

	req.Header.Set("User-Agent", "mcstats/1.0.0")
	req.Header.Set("Content-Type", "application/json")
	resp, err := httpClient.Do(req)

	if err != nil {
		return err
	}

	defer resp.Body.Close()

	return json.NewDecoder(resp.Body).Decode(v)
}
