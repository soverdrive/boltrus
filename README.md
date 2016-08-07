[Boltdb](https://github.com/boltdb/bolt) Hooks for [Logrus](https://github.com/Sirupsen/logrus) <img src="http://i.imgur.com/hTeVwmJ.png" width="40" height="40" alt=":walrus:" class="emoji" title=":walrus:"/>
-------

Hi, Boltrus is a BoltDB hooks for Logrus.

Install
-------
```shell
$ go get github.com/albert-widi/boltrus
```

Usage
------
```go
package main

import (
    "github.com/Sirupsen/logrus"
	"github.com/albert-widi/boltrus"
)

log := logrus.New()
hooker, err := boltrus.NewHook("files/logger/")

if err == nil {
  log.Hooks.Add(hooker)
}

log.WithFields(logrus.Fields{
	"name": "albert",
	"job":  "Awesome",
}).Error("Boltrus")
```

Info
----
Each log are separated into different db files.

Boltrus will make db files in the declared path. This is the db list:
* panic_log.db
* fatal_log.db
* error_log.db
* warning_log.db
* info_log.db
* debug_log.db
