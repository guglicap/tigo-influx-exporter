package main

import (
	"bytes"
	"flag"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"path"
	"strings"
	"time"
)

var DAQS_DIR = flag.String("daqs-dir", "/mnt/ffs/data/daqs", "directory where daqs file are stored.")
var INFLUX_URL = flag.String("influx-url", "", "influx api write endpoint")
var INFLUX_TOKEN = flag.String("t", "", "influx api token")
var INFLUX_ORG = flag.String("org", "", "influx organization")
var INFLUX_BUCKET = flag.String("b", "", "influx bucket")
var BUF_SIZE = flag.Int("buf-size", 350*37*150, "buffer size for the entire daq file.\nroughly set to n_entries * 150 * n_modules")
var IGNORE_FIELDS = flag.String("ignore-fields", "", "comma separated string of fields to ignore.\nes: RSSI,BRSSI,Status")

func main() {
	flag.Parse()

	lastUpdateFile := path.Join(*DAQS_DIR, ".last-update")
	var lastUpdate time.Time
	if f, err := os.Open(lastUpdateFile); err == nil {
		var s int64
		fmt.Fscanf(f, "%d", &s)
		lastUpdate = time.Unix(s, 0)
	} else {
		log("couldn't read last-update: ignore if this is the first run err:(%v)", err)
		lastUpdate = time.Unix(0, 0)
	}

	dir, err := os.ReadDir(*DAQS_DIR)
	if err != nil {
		panic("cannot read DAQS_DIR")
	}
	daqs := make([]string, 0)
	for _, f := range dir {
		if f.IsDir() || !strings.HasSuffix(f.Name(), ".csv") {
			continue
		}
		info, err := f.Info()
		if err != nil {
			continue
		}
		modTime := info.ModTime()
		// checking for at least a 2 seconds difference, as saving the time to last-update
		// as unix timestamp loses the nanoseconds and makes it updates the most recent daq again
		// even if no modifications has happened
		if modTime.Sub(lastUpdate) > 2*time.Second {
			daqs = append(daqs, f.Name())
			lastUpdate = modTime
		}
	}
	if len(daqs) == 0 {
		log("nothing to do, abort.")
		return
	}
	log("will update %v", daqs)
	daqsDir := os.DirFS(*DAQS_DIR)
	postFailed := false
	v := make([]byte, *BUF_SIZE)
	buf := bytes.NewBuffer(v)
	for _, fileName := range daqs {
		f, err := daqsDir.Open(fileName)
		if err != nil {
			log("error reading: %s, %v", fileName, err)
			continue
		}
		buf.Reset()
		err = encodeDaq(f, buf)
		f.Close()
		if err != nil {
			log("error encoding %s, skipping to next file, err: %v", fileName, err)
			continue
		}

		// DEBUG CODE //

		// _, err = io.Copy(os.Stdout, buf)
		// if err != nil {
		// 	log("cannot print lp: %v",)
		// }
		// continue
		

		postURL, err := url.Parse(*INFLUX_URL)
		if err != nil || len(*INFLUX_URL) == 0 {
			log("invalid or empty influx url, err:%v", err)
			return
		}
		params := url.Values{}
		params.Add("org", *INFLUX_ORG)
		params.Add("bucket", *INFLUX_BUCKET)
		params.Add("precision", "s")
		postURL.RawQuery = params.Encode()

		r, err := http.NewRequest(http.MethodPost, postURL.String(), buf)
		if err != nil {
			log("error creating POST request: %v", err)
			continue
		}
		r.Header.Add("Authorization", "Token "+*INFLUX_TOKEN)
		resp, err := http.DefaultClient.Do(r)
		if err != nil {
			log("error sending POST req: %v", err)
			postFailed = true
			continue
		}
		if resp.StatusCode != 204 {
			log("response status is not okay: %v", resp.StatusCode)
			if resp.StatusCode == http.StatusNotFound {
				log("http 404 while posting to: %s", postURL)
			}
			postFailed = true
			continue
		}
		log("successfully posted %v", fileName)
	}
	if !postFailed {
		s := fmt.Sprintf("%d", lastUpdate.Unix())
		err = os.WriteFile(lastUpdateFile, []byte(s), 0x600)
		if err != nil {
			log("error writing last-update: %v", err)
		}
		log("successfully written last-update: %v", lastUpdate)
	} else {
		log("post to influxdb failed, not writing last-update")
	}
}

func log(format string, v ...interface{}) {
	v = append([]interface{}{time.Now().Format(time.RFC3339)}, v...)
	fmt.Fprintf(os.Stderr, "%s: "+format+"\n", v...)
}
