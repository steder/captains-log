package main

import (
	"fmt"
	"net/http"
	"os"
	"io"
	"time"
	"io/ioutil"
	"encoding/json"
	"bytes"
	"github.com/davecgh/go-spew/spew"
	"math"
	"net"
	"log"
)

/*
  Documentation: http://192.168.0.135/docs/api/

  http://192.168.0.135/api/v1/printer

 */

const PRINTER_HOST string = "192.168.0.135"
const DEBUG bool = false
const FILEFORMAT = "snap_%05d.jpg"

func takeSnapshot(snapshotid int) error {
	snapshot, err := http.Get(fmt.Sprintf("http://%s:8080/?action=snapshot", PRINTER_HOST))
	if nerr, ok := err.(net.Error); ok && nerr.Temporary() {
		fmt.Printf("Skipping snapshot %d", snapshotid)
		return err
	}
	if err != nil {
		// Shut 'er down
		log.Fatalf("Unable to load the snapshot: %s", err)
	}
	defer snapshot.Body.Close()

	destPath := fmt.Sprintf(FILEFORMAT, snapshotid)
	destFile, err := os.Create(destPath)
	if err != nil {
		fmt.Printf("Unable to create %s: %s", destPath, err)
	}
	defer destFile.Close()
	os.Chmod(destPath, 0755)
	// TODO check errors here
	io.Copy(destFile, snapshot.Body)
	// TODO check more errors here

	return nil
}

func takeSnapshots(secondsRemaining float64) {
	fps := 30.0
	var hours float64 = secondsRemaining / 3600.0
	fmt.Printf("Hours: %v\n", hours)
	videoDuration := hours * 10.0
	if videoDuration < 10.0 {
		videoDuration = 10.0
	}
	fmt.Printf("Video Duration: %v\n", videoDuration)
	snapshotCount := fps * videoDuration
	delaySeconds := secondsRemaining / snapshotCount
	if delaySeconds < 2.0 {
		delaySeconds = 2.0
		snapshotCount = secondsRemaining / 2.0
	}
	snapshotDelay := time.Duration(delaySeconds * 1000.0 * 1000.0 * 1000.0)
	fmt.Printf("Capturing %f frames for a %fs video at %f FPS\n", snapshotCount, videoDuration, fps)
	fmt.Printf("Capturing frame every %f seconds for remaining %f seconds\n", snapshotDelay.Seconds(), secondsRemaining)
	// TODO: Replace with a select statement and a timer?
 	for i := 0; i < int(math.RoundToEven(snapshotCount)); {
		err := takeSnapshot(i)
		if err == nil {
			i++
		}
		time.Sleep(snapshotDelay)
	}
}

type Printer struct {
	Status string `json:"status""`
}

func checkAPI(path string, thing interface{}) {
	printerEndpoint := fmt.Sprintf("http://%s/api/v1/%s", PRINTER_HOST, path)
	printerInfo, err := http.Get(printerEndpoint)
	if err != nil {
		// handle it
		fmt.Println("Unable to load the endpoint: %s", err)
	}
	defer printerInfo.Body.Close()
	jsonBytes, err := ioutil.ReadAll(printerInfo.Body)
	if err != nil {
		fmt.Printf("Error reading response body; %s\n", err)
	}
	if DEBUG {
		var debugOutput bytes.Buffer
		json.Indent(&debugOutput, jsonBytes, "", "\t")
		fmt.Println(debugOutput.String())
	}
	err = json.Unmarshal(jsonBytes, thing)
	if err != nil{
		spew.Dump(thing)
		panic(err)
	}
}

/*func checkPrinterStatus() string {
	var p Printer;
	checkAPI("printer", &p)
	return p.Status
}*/

/*
/api/v1/print_job API response

 */

type PrintJob struct {
	//Cleaned *time.Time `json:"datetime_cleaned"`
	//Finished *time.Time `json:"datetime_finished"`
	Started string `json:"datetime_started"`
	Name string
	Progress float64
	ReprintUUID string `json:"reprint_original_uuid"`
	Result string
	Source string
	Application string `json:"source_application"`
	User string `json:"source_user"`
	State string
	TimeElapsed int `json:"time_elapsed"`
	TimeTotal int `json:"time_total"`
	Uuid string
}

func (j PrintJob) GetStartedTime() (time.Time, error) {
	return time.Parse("2006-01-02T15:04:05", j.Started)
}

func getPrintJob() PrintJob {
	var job PrintJob;
	checkAPI("print_job", &job)
	return job
}

func getPrintJobTimeRemaining() int {
	var job PrintJob
	checkAPI("print_job", &job)
	start, err := job.GetStartedTime()
	if err != nil {
		fmt.Printf("Couldn't parse Started time: %s", err)
	}
	fmt.Printf("Job started running at %v with %d seconds remaining", start, job.TimeTotal - job.TimeElapsed)
	return job.TimeTotal - job.TimeElapsed
}

func main() {
	fmt.Printf("Connecting to %s\n", PRINTER_HOST )
	for {
		job := getPrintJob()
		fmt.Printf("Checking printer status: %v ...\n", job)
		if job.State == "printing" && job.TimeTotal > 0 {
			fmt.Printf("\nStarting timelapse capture...      \n")
			timeRemaining := getPrintJobTimeRemaining()
			takeSnapshots(float64(timeRemaining))
			break
		}
		time.Sleep(5 * time.Second)
	}
	fmt.Printf("Done!\n")
}