package main

import (
	"encoding/csv"
	"io"
	"strings"
)

func encodeDaq(f io.Reader, buf io.StringWriter) error {
	r := csv.NewReader(f)
	r.ReuseRecord = true
	header, err := r.Read()
	if err != nil {
		return err
	}
	cols := make([]string, len(header))
	copy(cols, header) // needed because ReuseRecord == true

	ignoreFields := strings.Split(*IGNORE_FIELDS, ",")
	for rec, err := r.Read(); err == nil; rec, err = r.Read() {
		// loop over all fields, create a new line for every module
		curModule := ""
		fields := ""
		ts := ""
		field_loop:
		for i, fValue := range rec {
			colName := cols[i]
			if colName == "TimeStamp" {
				ts = strings.TrimSuffix(fValue, ".000000")
				continue
			}
			if !strings.HasPrefix(colName, "LMU") {
				continue
			}
			// fetch module and field name from LMU_<module>_<fieldName>
			module, fName := moduleField(colName)
			if curModule == "" {
				curModule = module
			}
			// flush previous module data
			if curModule != module {
				// if no data, we simply update the curModule variable and carry on
				if len(fields) == 0 {
					curModule = module
					continue
				}

				// manually creating line protocol here
				buf.WriteString("tigo,")
				buf.WriteString("module=" + curModule + " ")
				buf.WriteString(fields[:len(fields)-1] + " ")
				buf.WriteString(ts + "\n")
				
				// reset for next module
				fields = ""
				curModule = module
			}
			// check if field needs to be skipped
			for _, f := range ignoreFields {
				if fName == f {
					continue field_loop
				}
			}
			if len(fValue) == 0 {
				continue
			}
			fields += fName + "=" + fValue + ","
		}

		if len(fields) == 0 {
			continue
		}

		// flush last module values
		buf.WriteString("tigo,")
		buf.WriteString("module=" + curModule + " ")
		buf.WriteString(fields[:len(fields)-1] + " ")
		buf.WriteString(ts + "\n")
	}
	return err
}

// turns LMU_<module>_<fieldName> into (<module>, <fieldName>)
func moduleField(s string) (module string, field string) {
	x := strings.Split(s, "_")
	return x[1], x[2]
}
