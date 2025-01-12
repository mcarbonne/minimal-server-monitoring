package storage

import (
	"encoding/json"
	"os"

	"github.com/mcarbonne/minimal-server-monitoring/v2/pkg/logging"
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
		logging.Fatal("%v", err.Error())
	}
	js.database = decoded.Database
	if js.database == nil {
		js.database = make(map[string]string)
	}

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
		err = encoder.Encode(&toBeEncoded)
		if err != nil {
			logging.Fatal("Unable to encode: %v", err.Error())
		}
	}
	js.MemoryStorage.syncUnsafe(force)
}
