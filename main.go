package main

import (
	"fmt"
	"my-corn/myCron"
)

func main() {
	c := myCron.Cron{}
	cron := c.Create()
	des1 := "*/1 * * * * * *"
	des2 := "*/5 * * * * * *"
	_, err := cron.AddJob(des1, "job1")
	if err != nil {
		fmt.Println(err)
		return
	}

	_, err = cron.AddJob(des2, "job2")
	if err != nil {
		fmt.Println(err)
		return
	}
	go c.Start()
	defer c.StopCron()

}
