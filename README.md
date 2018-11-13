# go-cron
Rewrite linux-crontab by golang

## demo code
package main

import (
	"fmt"
	"../cron"
)

type DemoLogger struct{}

func (self *DemoLogger) Logger(format string, args ...interface{}) {
	//handle log
}

func ExecDemo() {
	fmt.Println("hello world")
}

//precision to second
var c = cron.New()

func main() {
	log := &DemoLogger{}
	//add job with log
	c.AddJob("*/1 * * * * *", ExecDemo, log)
	//add job with no log
	jobid, err := c.AddJob("*/1 * * * * *", ExecDemo, nil)
	c.Start(jobid)
	c.Stop(jobid)
	c.Delete(jobid)

	//When your program restarted and you maybe need rebuild you jobs
	//from config file or db you can use like this
  for _,job:=range GetJobs() {
     	execFunc:=makeFunc(job)
     	//c.AddJobWithJobID("*/1 * * * * *",ExecDemo,my_jobid,log)
	     id, _ := c.AddJobWithJobID(job.spec, execFunc, job.jobid, log)
	     c.Start(id)
  }
}
