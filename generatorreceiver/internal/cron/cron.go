package cron

import (
	"github.com/robfig/cron/v3"
	"log"
	"os"
)

var cronInstance *cron.Cron

func init() {
	cronInstance = cron.New(
		cron.WithLogger(
			cron.PrintfLogger(log.New(os.Stdout, "cron: ", log.LstdFlags))))
}

func Add(spec string, function func()) (cron.EntryID, error) {
	return cronInstance.AddFunc(spec, function)
}

func Start() {
	cronInstance.Start()
}

func Stop() {
	cronInstance.Stop()
}
