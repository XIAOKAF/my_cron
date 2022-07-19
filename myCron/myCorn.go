package myCron

import (
	"fmt"
	"github.com/gorhill/cronexpr"
	"log"
	"sync"
	"time"
)

type Entry struct {
	Ticker   *time.Ticker
	NextTime time.Time
	Duration time.Duration
	Name     string
	Id       int
}

type Cron struct {
	Entries   []*Entry
	RunningMu sync.Mutex
	Add       chan *Entry
	Del       chan *Entry
	Stop      chan struct{}
	Status    bool //标识是否为运行状态
	NextId    int
}

type options func(*Cron)

func (c *Cron) Create(opts ...options) *Cron {
	cron := &Cron{}
	for _, opt := range opts {
		opt(cron)
	}
	c.Status = false
	return cron
}

func (c *Cron) AddJob(des, name string) (int, error) {
	c.RunningMu.Lock()
	defer c.RunningMu.Unlock()

	c.NextId++
	now := time.Now()
	expression, err := cronexpr.Parse(des)
	if err != nil {
		log.Printf("解析cron表达式失败：%q", err)
		return c.NextId, err
	}

	duration := expression.Next(now).Sub(now)
	ticker := time.NewTicker(duration)

	newJob := &Entry{
		Id:       c.NextId,
		Ticker:   ticker,
		NextTime: expression.Next(time.Now()),
		Duration: duration,
		Name:     name,
	}

	if !c.Status {
		c.Entries = append(c.Entries, newJob)
	} else {
		c.Add <- newJob
	}
	return c.NextId, nil
}

func (c *Cron) DelJob(id int) {
	c.RunningMu.Lock()
	defer c.RunningMu.Unlock()
	jobRemoved := &Entry{
		Id: id,
	}
	if c.Status {
		c.Del <- jobRemoved
		return
	}
	var newEntries []*Entry
	for _, e := range c.Entries {
		if e.Id != id {
			newEntries = append(newEntries, e)
		}
	}
}

func (c *Cron) Start() {
	c.RunningMu.Lock()
	defer c.RunningMu.Unlock()
	if c.Status {
		return
	}
	c.Status = true
	go c.run()
}

func (c *Cron) Reset(id int, des string) {
	c.RunningMu.Lock()
	defer c.RunningMu.Unlock()
	for _, entryReset := range c.Entries {
		if entryReset.Id == id {
			now := time.Now()
			expression, err := cronexpr.Parse(des)
			if err != nil {
				log.Printf("解析cron表达式失败：%q", err)
				return
			}
			duration := expression.Next(now).Sub(now)
			entryReset.Ticker.Reset(duration)
		}
	}
}

func (c *Cron) StopCron() {
	c.RunningMu.Lock()
	defer c.RunningMu.Unlock()
	if c.Status {
		c.Stop <- struct{}{}
		c.Status = false
	}
}

func (c *Cron) run() {
	now := time.Now()
	for _, entry := range c.Entries {
		entry.NextTime = now.Add(entry.Duration)
	}
	for _, e := range c.Entries {
		for {
			select {
			case <-e.Ticker.C:
				fmt.Printf("%s:\n", e.Name)
				fmt.Printf("%s\n", <-e.Ticker.C)
			case newEntry := <-c.Add:
				c.Entries = append(c.Entries, newEntry)
				c.NextId = newEntry.Id
			case entryRemoved := <-c.Del:
				var newEntries []*Entry
				for _, e := range c.Entries {
					if e.Id != entryRemoved.Id {
						newEntries = append(newEntries, e)
					}
				}
			case <-c.Stop:
				fmt.Println("Cron has stopped.")
				return
			}
		}
	}
}
