package storage

import (
	"encoding/json"
	"os"

	"github.com/mcarbonne/minimal-server-monitoring/pkg/logging"
)

type JsonStorageFile struct {
	Database map[string]string `json:"database"`
}

type JSONStorage struct {
	MemoryStorage
	jsonPath string
}

func NewJSONStorage(jsonPath string) Storager {
	js := &JSONStorage{jsonPath: jsonPath,
		MemoryStorage: MemoryStorage{database: make(map[string]string)}}

	storageFile, err := os.Open(jsonPath)

	if err != nil {
		logging.Warning("Unable to load configuration: %v", err)
		return js
	} else {
		defer storageFile.Close()
	}

	jsonParser := json.NewDecoder(storageFile)
	jsonParser.DisallowUnknownFields()
	decoded := JsonStorageFile{}
	err = jsonParser.Decode(&decoded)
	if err != nil {
		logging.Fatal(err.Error())
	}
	js.database = decoded.Database

	return js
}

func (js *JSONStorage) Sync(force bool) {
	js.mutex.Lock()
	defer js.mutex.Unlock()
	if js.unsyncedChanges || force {
		toBeEncoded := JsonStorageFile{
			Database: js.database,
		}
		storageFile, err := os.Create(js.jsonPath)
		if err != nil {
			logging.Fatal("Unable to open: %v", err.Error())
		} else {
			defer storageFile.Close()
		}
		encoder := json.NewEncoder(storageFile)
		encoder.SetIndent("", " ")
		encoder.Encode(&toBeEncoded)
	}
	js.MemoryStorage.syncUnsafe(force)
}
