package main

import (
	"bytes"
	"encoding/csv"
	"io"
	"strings"
)

func encodeDaq(f io.Reader, buf *bytes.Buffer) (error) {
	r := csv.NewReader(f)
	r.ReuseRecord = true
	header, err := r.Read()
	if err != nil {
		return err
	}
	cols := make([]string, len(header))
	copy(cols, header) // needed because ReuseRecord == true
	for rec, err := r.Read(); err == nil; rec, err = r.Read() {
		// loop over all fields, create a new line for every module
		curModule := ""
		fields := ""
		ts := ""
		for i, fValue := range rec {
			fName := cols[i]
			if fName == "TimeStamp" {
				ts = strings.TrimSuffix(fValue, ".000000")
				continue
			}
			if !strings.HasPrefix(fName, "LMU") {
				continue
			}
			module, vName := moduleField(fName)
			if curModule == "" {
				curModule = module
			} else if curModule != module {
				if len(fields) == 0 {
					curModule = module
					continue
				}
				// flush current module values
				buf.WriteString("tigo,")
				buf.WriteString("module=" + curModule + " ")
				buf.WriteString(fields[:len(fields)-1] + " ")
				buf.WriteString(ts + "\n")
				fields = ""
				curModule = module
			}
			// select fields to save
			switch vName {
			case "Status":
				fallthrough
			case "Flags":
				fallthrough
			case "RSSI":
				fallthrough
			case "ID":
				fallthrough
			case "Details":
				fallthrough
			case "BRSSI":
				continue
			}
			if len(fValue) == 0 {
				continue
			}
			fields += vName + "=" + fValue + ","
		}
	}
	return nil
}

func moduleField(s string) (module string, field string) {
	x := strings.Split(s, "_")
	return x[1], x[2]
}
