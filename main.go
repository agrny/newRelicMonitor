// Package newrelicmonitor
package main

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"new-relic-monitor/config"
)

type BucketFile struct {
	ModifedAt time.Time
	Date      string
	Time      string
	Size      int
	Name      string
}

const configFilePath = "./env.json"

func main() {
	configOptions := config.Config{}

	configOptions.ReadConfigFile(configFilePath)

	bucketURL := fmt.Sprintf("s3://%s", configOptions.BucketName)
	bufOut := bytes.Buffer{}
	bufErr := bytes.Buffer{}

	cmd := exec.Command("aws", "s3", "ls", bucketURL)
	cmd.Stdout = &bufOut
	cmd.Stderr = &bufErr
	cmd.Stdin = os.Stdin

	err := cmd.Run()
	if err != nil {
		fmt.Printf("ERROR RUNNING PROCESS\n%s\n", bufErr.String())
		os.Exit(1)
	}

	outputRawString := strings.TrimSpace(bufOut.String())
	lines := strings.Split(outputRawString, "\n")

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
		if backupOccurred(bFile.ModifedAt) {
			fmt.Printf("Backup Occurred: %s\n", bFile.Name)
		}
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
		Name:      name,
		ModifedAt: lastModified,
		Size:      size,
		Date:      date,
		Time:      clockTime,
	}
	return toAdd, nil
}

func backupOccurred(modifiedAtTime time.Time) bool {
	today := time.Now().Local()
	modifiedAt := modifiedAtTime.Local()
	fmt.Printf("today: %v  modifiedAt %v\n", today, modifiedAt)
	cutoff := time.Now().UTC().Add(-24 * time.Hour)
	return modifiedAt.After(cutoff)
}
