// Package newrelicmonitor
package main

import (
	"bytes"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"new-relic-monitor/config"
	"new-relic-monitor/newRelic"
)

type BucketFile struct {
	ModifiedAt time.Time
	Date       string
	Time       string
	Size       int
	Name       string
}

func (bf *BucketFile) ToNewRelicEvent() map[string]any {
	event := map[string]any{
		"name":       bf.Name,
		"modifiedat": bf.ModifiedAt.String(),
		"size":       bf.Size,
	}

	return event
}

const configFilePath = "./config.json"

func main() {
	configPath := flag.String("config", "", "path to config file")
	flag.Parse()

	if *configPath == "" {
		*configPath = os.Getenv("S3BACKUPCONFIG")
		if *configPath == "" {
			*configPath = configFilePath
		}
	}
	configOptions := config.Config{}
	configOptions.Load(*configPath)

	newRelicApp, err := newrelic.NewApplicationWithConfig(configOptions)
	if err != nil {
		fmt.Printf("ERROR CREATING NEW RELIC APP\n%s\n", err)
		os.Exit(1)
	}
	defer newRelicApp.Shutdown(10 * time.Second)

	lines, err := listBucket(configOptions.BucketName)
	if err != nil {
		fmt.Printf("ERROR LISTING BUCKET\n%s\n", err)
		os.Exit(1)
	}

	bucketFiles := []BucketFile{}

	for _, line := range lines {
		if len(line) > 0 {
			toAdd, err := populateFileFromLine(line, configOptions)
			if err != nil {
				fmt.Printf("error populating file: %v", err)
				os.Exit(1)
			}

			bucketFiles = append(bucketFiles, toAdd)

		} else {
			println("skipping empty line")
		}
	}

	for _, bFile := range bucketFiles {
		if backupOccurredWithinDay(bFile.ModifiedAt) {
			fmt.Printf("Backup Occurred: %s\n", bFile.Name)
		}
		newRelicApp.RecordCustomEvent("S3BACKUPCHECK", bFile.ToNewRelicEvent())
	}
}

func populateFileFromLine(cmdOutput string, config config.Config) (BucketFile, error) {
	columns := strings.Fields(cmdOutput)
	if len(columns) < 4 {
		return BucketFile{}, fmt.Errorf("unexpected line format: %s", cmdOutput)
	}
	date := columns[0]
	clockTime := columns[1]
	size, _ := strconv.Atoi(columns[2])

	// handles if filename has spaces
	name := strings.Join(columns[3:], " ")

	// fmt.Printf("date: %s time: %s size: %d name: %s\n", date, clockTime, size, name)
	lastModified, err := time.ParseInLocation(config.TimeParseLayout, fmt.Sprintf("%s %s", date, clockTime), time.UTC)
	if err != nil {
		return BucketFile{}, err
	}
	toAdd := BucketFile{
		Name:       name,
		ModifiedAt: lastModified,
		Size:       size,
		Date:       date,
		Time:       clockTime,
	}
	return toAdd, nil
}

func backupOccurredWithinDay(modifiedAtTime time.Time) bool {
	today := time.Now().Local()
	modifiedAt := modifiedAtTime.Local()
	fmt.Printf("today: %v  modifiedAt %v\n", today, modifiedAt)
	cutoff := time.Now().UTC().Add(-24 * time.Hour)
	return modifiedAt.After(cutoff)
}

func listBucket(bucketName string) ([]string, error) {
	bucketURL := fmt.Sprintf("s3://%s", bucketName)
	cmd := exec.Command("aws", "s3", "ls", bucketURL)
	var bufOut, bufErr bytes.Buffer
	cmd.Stdout = &bufOut
	cmd.Stderr = &bufErr
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("ERROR RUNNING PROCESS\n%s", bufErr.String())
	}
	lines := strings.Split(strings.TrimSpace(bufOut.String()), "\n")
	return lines, nil
}
