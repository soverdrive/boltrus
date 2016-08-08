package boltrus

import (
	"time"

	"encoding/json"

	"github.com/Sirupsen/logrus"
	"github.com/boltdb/bolt"
)

//Hooker for hook
type Hooker struct {
	DBPath  string
	BoltMap map[string]*bolt.DB
}

var panicdb string = "panic_log.db"
var fataldb string = "fatal_log.db"
var errdb string = "err_log.db"
var warndb string = "warning_log.db"
var infodb string = "info_log.db"
var debugdb string = "debug_log.db"

//NewHook add new hook for boltdb
func NewHook(path string) (*Hooker, error) {
	boltHook := &Hooker{
		DBPath: path,
		BoltMap: map[string]*bolt.DB{
			"panic":   openDB(path, panicdb),
			"fatal":   openDB(path, fataldb),
			"error":   openDB(path, errdb),
			"warning": openDB(path, warndb),
			"info":    openDB(path, infodb),
			"debug":   openDB(path, debugdb),
		},
	}

	return boltHook, nil
}

func openDB(fullPath string, dbName string) *bolt.DB {
	db, err := bolt.Open(fullPath+dbName, 0600, nil)
	if err != nil {
		logrus.Errorf("Cannot open %s database", dbName)
	}

	return db
}

type logrusStash struct {
	Type        string      `json:"type"`
	TimeStamp   string      `json:"timestamp"`
	Sourcehost  string      `json:"host"`
	Message     string      `json:"message"`
	Fields      interface{} `json:"fields"`
	Application string      `json:"application"`
	File        string      `json:"file"`
	Level       string      `json:"level"`
}

//Fire logging to hook
func (bhook *Hooker) Fire(entry *logrus.Entry) error {
	stash := logrusStash{}

	stash.Message = entry.Message
	stash.TimeStamp = entry.Time.UTC().Format(time.RFC3339Nano)
	stash.Level = entry.Level.String()
	stash.Fields = entry.Data

	messageByte := []byte(stash.Message)
	keyTime := []byte(stash.TimeStamp)
	//convert fields to json
	dataByte, _ := json.Marshal(entry.Data)
	dataLength := len(entry.Data)

	tx, _ := bhook.BoltMap[stash.Level].Begin(true)
	tx.CreateBucketIfNotExists(messageByte)
	//this is message log message bucket
	bucket := tx.Bucket(messageByte)

	putFields(bucket, keyTime, dataByte, dataLength)
	tx.Commit()

	return nil
}

//Levels rerturn available levels in hook
func (bhook *Hooker) Levels() []logrus.Level {
	return []logrus.Level{
		logrus.PanicLevel,
		logrus.FatalLevel,
		logrus.ErrorLevel,
		logrus.WarnLevel,
		logrus.InfoLevel,
		logrus.DebugLevel,
	}
}

func putFields(bucket *bolt.Bucket, key []byte, bucketName []byte, dataLength int) {
	var dataValue []byte

	if dataLength > 0 {
		dataValue = bucketName
	} else {
		dataValue = []byte("no_fields")
	}

	//this is fields bucket
	bucket.CreateBucketIfNotExists(dataValue)
	fieldsBucket := bucket.Bucket(dataValue)
	fieldsBucket.Put(key, []byte("1"))
}

//GetPanicList return all panics available in panic boltdb
func (bhook *Hooker) GetPanicList() []string {
	var list []string

	tx, _ := bhook.BoltMap["panic"].Begin(true)
	tx.ForEach(func(name []byte, _ *bolt.Bucket) error {
		list = append(list, string(name))
		return nil
	})
	tx.Commit()

	return list
}

//GetFatalList return all fatals available in fatal boltdb
func (bhook *Hooker) GetFatalList() []string {
	var list []string

	tx, _ := bhook.BoltMap["fatal"].Begin(true)
	tx.ForEach(func(name []byte, _ *bolt.Bucket) error {
		list = append(list, string(name))
		return nil
	})
	tx.Commit()

	return list
}

//GetErrorList return all errors available in error boltdb
func (bhook *Hooker) GetErrorList() []string {
	var list []string

	tx, _ := bhook.BoltMap["error"].Begin(true)
	tx.ForEach(func(name []byte, _ *bolt.Bucket) error {
		list = append(list, string(name))
		return nil
	})
	tx.Commit()

	return list
}

//GetWarningList return all warnings available in warning boltdb
func (bhook *Hooker) GetWarningList() []string {
	var list []string

	tx, _ := bhook.BoltMap["warning"].Begin(true)
	tx.ForEach(func(name []byte, _ *bolt.Bucket) error {
		list = append(list, string(name))
		return nil
	})
	tx.Commit()

	return list
}

//GetInfoList return all info available in info boltdb
func (bhook *Hooker) GetInfoList() []string {
	var list []string

	tx, _ := bhook.BoltMap["info"].Begin(true)
	tx.ForEach(func(name []byte, _ *bolt.Bucket) error {
		list = append(list, string(name))
		return nil
	})
	tx.Commit()

	return list
}

//GetDebugList return all debugs available in debug boltdb
func (bhook *Hooker) GetDebugList() []string {
	var list []string

	tx, _ := bhook.BoltMap["debug"].Begin(true)
	tx.ForEach(func(name []byte, _ *bolt.Bucket) error {
		list = append(list, string(name))
		return nil
	})
	tx.Commit()

	return list
}
