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
)

type BucketFile struct {
	ModifedAt time.Time
	Date      string
	Time      string
	Size      int
	Name      string
}

const (
	TimeParseLayout = "2006-01-02 15:04:05"
	BucketName      = "backup-bucket-767398055203-us-east-1-an"
)

func main() {
	bucketURL := fmt.Sprintf("s3://%s", BucketName)
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
			toAdd, err := populateFileFromLine(line)
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

func populateFileFromLine(cmdOutput string) (BucketFile, error) {
	columns := strings.Split(cmdOutput, " ")
	date := columns[0]
	clockTime := columns[1]
	size, _ := strconv.Atoi(columns[10])
	name := columns[11]
	fmt.Printf("date: %s time: %s size: %d name: %s\n", date, clockTime, size, name)
	println()
	lastModified, err := time.ParseInLocation(TimeParseLayout, fmt.Sprintf("%s %s", date, clockTime), time.UTC)
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
	// fmt.Printf("today: %v, modifiedDay %v\n", today, modifedAtDay)
	fmt.Printf("modifiedDayTime %v\n", modifiedAt)
	fmt.Printf("today %v\n", today)
	return modifiedAt.After(today.Add(-(24 * time.Hour)))
}
