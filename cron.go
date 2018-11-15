package cron

import (
	"time"
	"runtime"
)

//go version >= 1.9
type jobId = string

type Logger interface {
	Logger(format string, args ...interface{})
}

type FuncJob func()

func (f FuncJob) Run() { f() }

type cron struct {
	ticker *time.Ticker
	jobid  jobId
	add    chan Job
	del    chan jobId
	stop   chan jobId
	start  chan jobId
	jobs   map[jobId]Job
}

type Job interface {
	run()
	stop()
	start()
	delete()
	next(tick) bool
}

type schedule struct {
	second  map[int]struct{}
	minute  map[int]struct{}
	hour    map[int]struct{}
	day     map[int]struct{}
	month   map[int]struct{}
	week    map[int]struct{}
	job     FuncJob
	logger  Logger
	running bool
	jobId   string
}

type tick struct {
	second int
	minute int
	hour   int
	day    int
	month  int
	week   int
}

func New() *cron {
	c := &cron{
		ticker: time.NewTicker(time.Second),
		add:    make(chan Job, 1),
		del:    make(chan jobId),
		stop:   make(chan jobId),
		start:  make(chan jobId),
		jobs:   make(map[jobId]Job),
	}
	go func() {
		for {
			select {
			case t := <-c.ticker.C:
				c.schedule(t)

			case job := <-c.add:
				c.jobs[c.jobid] = job

			case jobid := <-c.stop:
				if schedule, ok := c.jobs[jobid]; ok {
					schedule.stop()
				}

			case jobid := <-c.start:
				if schedule, ok := c.jobs[jobid]; ok {
					schedule.start()
				}

			case jobid := <-c.del:
				if _, ok := c.jobs[jobid]; ok {
					delete(c.jobs, jobid)
				}
			}
		}
	}()
	return c
}

func (self *cron) schedule(t time.Time) {
	tick := currentTick(t)
	for _, job := range self.jobs {
		if job.next(tick) {
			go job.run()
		}
	}
}

func (self *cron) addJob(spec string, cmd func(), id jobId, logger Logger) (jobId, error) {
	schedule, err := parse(spec)
	if err != nil {
		return "", err
	}
	schedule.job = cmd
	self.jobid = id
	self.add <- schedule
	return id, nil
}

func (self *cron) AddJob(spec string, cmd func(), logger Logger) (jobId, error) {
	return self.addJob(spec, cmd, UUID(), logger)
}

func (self *cron) AddJobWithJobID(spec string, cmd func(), id jobId, logger Logger) (jobId, error) {
	return self.addJob(spec, cmd, id, logger)
}

func (self *cron) Start(id jobId) {
	self.start <- id
}

func (self *cron) Stop(id jobId) {
	self.stop <- id
}

func (self *cron) Delete(id jobId) {
	self.del <- id
}

func (self *schedule) run() {
	defer func() {
		if r := recover(); r != nil {
			const size = 64 << 10
			buf := make([]byte, size)
			buf = buf[:runtime.Stack(buf, false)]
			self.logf("Job(jobID:%s) run panic: %v\n%s", self.jobId, r, buf)
		}
	}()
	if self.running {
		self.job.Run()
	}
}

func (self *schedule) start() {
	self.running = true
}

func (self *schedule) stop() {
	self.running = false
}

//just for implements interfac
func (self *schedule) delete() {}

func (self *schedule) next(t tick) bool {
	if _, ok := self.second[t.second]; !ok {
		return false
	}

	if _, ok := self.minute[t.minute]; !ok {
		return false
	}

	if _, ok := self.hour[t.hour]; !ok {
		return false
	}

	_, day := self.day[t.day]
	_, week := self.week[t.week]
	if !day && !week {
		return false
	}

	if _, ok := self.month[t.month]; !ok {
		return false
	}

	return true
}

func (self *schedule) logf(format string, args ...interface{}) {
	if self.logger != nil {
		self.logger.Logger(format, args...)
	}
}

func currentTick(t time.Time) tick {
	return tick{
		second: t.Second(),
		minute: t.Minute(),
		hour:   t.Hour(),
		day:    t.Day(),
		month:  int(t.Month()),
		week:   int(t.Weekday()),
	}
}
