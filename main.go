// Package newrelicmonitor
package main

import (
	"bytes"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"slices"
	"strconv"
	"strings"
	"time"

	newrelic "new-relic-monitor/newRelic"
)

type BucketFile struct {
	ModifiedAt time.Time
	Date       string
	Time       string
	Size       int
	Name       string
}

func (bf *BucketFile) ToNewRelicEvent(backupSucceeded bool) map[string]any {
	backupStatus := 0
	if backupSucceeded {
		backupStatus = 1
	}

	event := map[string]any{
		"eventType":  "S3BACKUPCHECK",
		"name":       bf.Name,
		"modifiedat": bf.ModifiedAt.String(),
		"size":       bf.Size,
		"isBackedUp": backupStatus,
	}

	return event
}

const configFilePath = "./env.json"

func main() {
	configPath := flag.String("config", "", "path to config file")
	flag.Parse()

	if *configPath == "" {
		*configPath = os.Getenv("S3BACKUPCONFIG")
		if *configPath == "" {
			*configPath = configFilePath
		}
	}

	cfg, err := newrelic.LoadConfig(*configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR LOADING CONFIG: %s", err)
		os.Exit(1)
	}

	newRelicClient, err := newrelic.NewClientWithConfig(*cfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR CREATING NEW RELIC APP: %s", err)
		os.Exit(1)
	}

	lines, err := listBucket(cfg.BucketName)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR LISTING BUCKET: %s", err)
		os.Exit(1)
	}

	bucketFiles, err := populateBucketFiles(lines, *cfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error populating bucket files: %v", err)
		os.Exit(1)
	}

	mostRecent := slices.MaxFunc(bucketFiles, func(a, b BucketFile) int {
		return a.ModifiedAt.Compare(b.ModifiedAt)
	})

	backupStatus := backupOccurredWithinDay(mostRecent.ModifiedAt)
	newRelicClient.RecordCustomEvent("S3BACKUPCHECK", mostRecent.ToNewRelicEvent(backupStatus))
}

func populateBucketFiles(lines []string, cfg newrelic.Config) ([]BucketFile, error) {
	bucketFiles := []BucketFile{}
	for _, line := range lines {
		if len(line) == 0 {
			println("skipping empty line")
			continue
		}
		toAdd, err := populateFileFromLine(line, cfg)
		if err != nil {
			return nil, fmt.Errorf("error populating file: %w", err)
		}
		bucketFiles = append(bucketFiles, toAdd)
	}
	return bucketFiles, nil
}

func populateFileFromLine(cmdOutput string, config newrelic.Config) (BucketFile, error) {
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
	modifiedAt := modifiedAtTime.Local()
	cutoff := time.Now().UTC().Add(-24 * time.Hour)
	return modifiedAt.After(cutoff)
}

func listBucket(bucketName string) ([]string, error) {
	if _, err := exec.LookPath("aws"); err != nil {
		return nil, fmt.Errorf("aws cli not found in PATH")
	}
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
