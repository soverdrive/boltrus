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

const (
	panicdb = "panic_log.db"
	fataldb = "fatal_log.db"
	errdb   = "error_log.db"
	warndb  = "warning_log.db"
	infodb  = "info_log.db"
	debugdb = "debug_log.db"
)

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
	Date        string      `json:"date"`
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
	stash.TimeStamp = entry.Time.Format(time.RFC3339Nano)
	stash.Date = entry.Time.Format("02 Jan 2006")
	stash.Level = entry.Level.String()
	stash.Fields = entry.Data

	messageByte := []byte(stash.Message)
	keyTime := []byte(stash.TimeStamp)
	keyDate := []byte(stash.Date)
	//convert fields to json
	dataByte, _ := json.Marshal(entry.Data)
	dataLength := len(entry.Data)

	tx, _ := bhook.BoltMap[stash.Level].Begin(true)
	//create bucket for date
	tx.CreateBucketIfNotExists(keyDate)
	//accessing the date bucket
	dateBucket := tx.Bucket(keyDate)
	//create bucket for log message
	dateBucket.CreateBucketIfNotExists(messageByte)
	//accessing message bucket
	messageBucket := dateBucket.Bucket(messageByte)

	putFields(messageBucket, keyTime, dataByte, dataLength)
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

//LogType return available log tag in hook
func LogType() []string {
	return []string{
		"panic",
		"fatal",
		"error",
		"warning",
		"info",
		"debug",
	}
}

func putFields(bucket *bolt.Bucket, key []byte, bucketName []byte, dataLength int) {
	var dataValue []byte

	if dataLength > 0 {
		dataValue = bucketName
	} else {
		dataValue = []byte("no_fields")
	}

	//create fields bucket
	bucket.CreateBucketIfNotExists(dataValue)
	//accessing fields bucket
	fieldsBucket := bucket.Bucket(dataValue)
	//put timestamp key and value into fields bucket
	fieldsBucket.Put(key, []byte("1"))
}

//GetLogDate return all date available in a log db
func (bhook *Hooker) GetLogDate(db string) ([]string, error) {
	var list []string

	tx, err := bhook.BoltMap[db].Begin(true)

	if err != nil {
		return list, err
	}

	tx.ForEach(func(name []byte, _ *bolt.Bucket) error {
		list = append(list, string(name))
		return nil
	})
	tx.Commit()

	return list, err
}

//GetLogList return all error in a log db
func (bhook *Hooker) GetLogList(db string, date string) ([]string, error) {
	var list []string

	tx, err := bhook.BoltMap[db].Begin(true)

	if err != nil {
		return list, err
	}

	dateBucket := tx.Bucket([]byte(date))

	dateBucket.Tx().ForEach(func(name []byte, _ *bolt.Bucket) error {
		list = append(list, string(name))
		return nil
	})
	dateBucket.Tx().Commit()

	tx.Commit()

	return list, err
}

//GetLogFieldList return all fields in a log db
func (bhook *Hooker) GetLogFieldList(db string, date string, message string) (map[string][]string, error) {
	list := make(map[string][]string)

	tx, err := bhook.BoltMap[db].Begin(true)

	if err != nil {
		return list, err
	}

	dateBucket := tx.Bucket([]byte(date))
	messageBucket := dateBucket.Bucket([]byte(message))

	messageBucket.Tx().ForEach(func(fields []byte, fieldsBucket *bolt.Bucket) error {
		fieldsString := string(fields)
		fieldsBucket.ForEach(func(key []byte, value []byte) error {
			list[fieldsString] = append(list[fieldsString], string(key))
			return nil
		})
		return nil
	})
	messageBucket.Tx().Commit()
	tx.Commit()

	return list, err
}

//DeleteLog will delete the log in database in the given time (Deleting date bucket)
func (bhook *Hooker) DeleteLog(days int) {
	go scanDelete("panic", days, bhook)
	go scanDelete("fatal", days, bhook)
	go scanDelete("error", days, bhook)
	go scanDelete("warning", days, bhook)
	go scanDelete("info", days, bhook)
	go scanDelete("debug", days, bhook)
}

func scanDelete(db string, days int, bhook *Hooker) {
	tx, _ := bhook.BoltMap[db].Begin(true)

	tx.ForEach(func(name []byte, bucket *bolt.Bucket) error {
		logDate, _ := time.Parse("02 Jan 2006", string(name))

		if logDate.Add(time.Hour * 24 * time.Duration(days)).After(time.Now()) {
			tx.DeleteBucket(name)
		}
		return nil
	})
	tx.Commit()
}

//Dump will create all dbs text file of the log
func (bhook *Hooker) Dump(path string) error {
	return nil
}
