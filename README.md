## Tigo Influx Exporter

Should be run on Tigo CCA, uploads module data to an InfluxDB v2 server.

Cross compile with `GOARCH=arm GOARM=7 go build`, then copy to the CCA. 

A `cmd/fileserver.go` script makes it easy to copy the binary in case the CCA isn't setup with ftp or similar, simply run `go run cmd/fileserver.go` which will serve the **current directory** at `http:<your_ip>:8080`, you can then run `wget http://<your_ip>:8080/tigo-influx-exporter` on the CCA to download the file. Don't forget to `chmod +x` the binary.

Once on the CCA, you can set up a cronjob to run periodically. I would recommend a minimum interval of 3 min, as it might take a while to complete large uploads, and I haven't yet implemented a lock mechanism, so it's possible that the script might be started again while the previous one is still running. This shouldn't be an issue, it's just a waste of resources.

```
Usage of ./tigo-influx-exporter:
  -b string
        influx bucket
  -buf-size int
        buffer size for the entire daq file.
        roughly set to n_entries * 150 * n_modules (default 1942500)
  -daqs-dir string
        directory where daqs file are stored. (default "/mnt/ffs/data/daqs")
  -ignore-fields string
        comma separated string of fields to ignore.
        es: RSSI,BRSSI,Status
  -influx-url string
        influx api write endpoint
  -org string
        influx organization
  -t string
        influx api token
```
