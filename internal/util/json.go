package util

import (
	"encoding/json"
	"os"
)

func LoadJSON(filename string, object interface{}) (err error) {

	file, err := os.Open(filename)
	if err != nil {
		return
	}

	dec := json.NewDecoder(file)
	err = dec.Decode(object)
	if err != nil {
		return
	}

	file.Close()
	return
}

func SaveJSON(filename string, object interface{}) (err error) {

	file, err := os.Create(filename)
	if err != nil {
		return
	}

	enc := json.NewEncoder(file)
	err = enc.Encode(object)
	if err != nil {
		return
	}

	file.Close()
	return
}
