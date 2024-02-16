## Tigo Influx Exporter

Should be run on Tigo CCA, uploads module data to an InfluxDB v2 server.

**Note:** by default, it ignores the `Status`,`Flags`,`RSSI`,`ID`,`Details`,`BRSSI` fields for every module to save on data since they weren't needed in my use case. This should be easy to change by modifying the code [here](https://github.com/guglicap/tigo-influx-exporter/blob/feefef21e06dbbb60750c6904618b3f63f946b08/daq2lp.go#L50).

```
Usage of hacking/tigo/tigo-influx-exporter:
  -b string
    	influx bucket
  -daqs-dir string
    	directory where daqs file are stored. (default "/mnt/ffs/data/daqs")
  -influx-url string
    	influx api write endpoint
  -org string
    	influx organization
  -t string
    	influx api token
```
