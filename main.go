package main

import (
	"flag"
	"log"
	"time"

	"zd/config"
)

func main() {

	// default config is zd.yaml in the path
	filename := flag.String("c", "zd.yaml", "location of the configuration file")
	flag.Parse()

	for {
		cfg, err := config.Load(filename)
		if err != nil {
			log.Printf("%s\n", err)

		} else {

			var scrubTasks, snapshotTasks, replicationTasks []config.Task
			for _, t := range cfg.Tasks {

				if t.IsNow() {
					switch t.Task {
					case "scrub":
						scrubTasks = append(scrubTasks, t)
					case "snapshot":
						snapshotTasks = append(snapshotTasks, t)
					case "replication", "replicate":
						replicationTasks = append(replicationTasks, t)
					case "snapshot and replication", "snapshot and replicate":
						t1 := t
						t1.Task = "snapshot"

						t2 := t
						t2.Task = "replication"

						snapshotTasks = append(snapshotTasks, t1)
						replicationTasks = append(replicationTasks, t2)
					}
				}
			}

			for _, t := range snapshotTasks {
				t.Run()
			}

			for _, t := range replicationTasks {
				t.Run()
			}

			for _, t := range scrubTasks {
				t.Run()
			}
		}

		// Sleep until next minute
		now := time.Now().UTC()
		next := time.Date(now.Year(), now.Month(), now.Day(), now.Hour(), now.Minute()+1, 0, 0, time.UTC)
		time.Sleep(next.Sub(now))
	}

}
