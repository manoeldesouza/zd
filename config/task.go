package config

import (
	"fmt"
	"log"
	"sort"
	"strings"
	"time"

	"zd/zfs"
)

type Task struct {
	Task      string   `yaml:"task"`
	Target    string   `yaml:"target"`
	Frequency string   `yaml:"frequency"`
	When      string   `yaml:"when"`
	Retention int      `yaml:"retention"`
	Recursive bool     `yaml:"recursive"`
	Raw       bool     `yaml:"raw"`
	To        []string `yaml:"to"`
}

func (t *Task) IsNow() bool {

	now := time.Now().UTC()
	switch t.Frequency {
	case "hourly":
		if now.Format("04") == t.When {
			return true
		}

	case "daily":
		if now.Format("15:04") == t.When {
			return true
		}

	case "weekly":
		if now.Format("Mon 15:04") == t.When {
			return true
		}

	case "monthly":
		if now.Format("01 15:04") == t.When {
			return true
		}
	}

	return false
}

func (t *Task) exists() bool {

	if t.Task == "scrub" {
		return zfs.Exists("", "pool", t.Target)
	}

	return zfs.Exists("", "", t.Target)
}

func (t *Task) Run() /* error */ {

	switch t.Task {
	case "scrub":
		t.scrub()

	case "snapshot":
		t.snapshot()

	case "replication", "replicate":
		t.replicate()

		// case "snapshot and replication", "snapshot and replicate":
		// 	t.snapshot()
		// 	t.replicate()
	}
}

func (t *Task) scrub() {
	err := zfs.Scrub(t.Target)
	if err != nil {
		log.Printf("%s\n", err)
		return
	}
}

func (t *Task) snapshot() {

	single_snapshot(t.Target, t.Frequency, t.Retention)

	l := zfs.ListChildsNamesOf(t.Target)
	if t.Recursive && len(l) > 0 {
		for _, d := range l {
			single_snapshot(d, t.Frequency, t.Retention)
		}
	}
}

func single_snapshot(datasetName, frequency string, retention int) {

	now := time.Now().UTC()
	snapshotName := fmt.Sprintf("%s_%s_%s", "auto", now.Format("2006-01-02_15-04-05_UTC"), frequency)
	err := zfs.Snapshot(datasetName, snapshotName)
	if err != nil {
		log.Printf("%s\n", err)
		return
	}
	log.Printf("Ok: snapshoted dataset %s as %s", datasetName, snapshotName)

	suffix := frequency
	l, err := zfs.Prune(datasetName, suffix, retention)
	if err != nil {
		log.Printf("run error: %s\n", err)
		return
	}
	if len(l) > 1 {
		log.Printf("Ok: pruned dataset %s removing %d snapshots", datasetName, len(l))
	} else if len(l) == 1 {
		log.Printf("Ok: pruned dataset %s removing %s", datasetName, l[0])
	}
}

func (t *Task) replicate() {

	single_replication(t.Target, t.Frequency, t.Raw, t.To)

	l := zfs.ListChildsNamesOf(t.Target)
	if t.Recursive && len(l) > 0 {
		for _, d := range l {

			appendix := strings.Replace(d, t.Target, "", -1)
			newTo := []string{}
			for i := range t.To {

				if strings.Contains(t.To[i], "@") {
					dataset := strings.Split(t.To[i], "@")[0]
					host := strings.Split(t.To[i], "@")[1]
					to := fmt.Sprintf("%s%s@%s", dataset, appendix, host)
					newTo = append(newTo, to)

				} else {
					to := fmt.Sprintf("%s%s", t.To[i], appendix)
					newTo = append(newTo, to)
				}
			}
			single_replication(d, t.Frequency, t.Raw, newTo)
		}
	}
}

func single_replication(datasetName, frequency string, raw bool, to []string) {

	sourceSnapshots := zfs.ListSnapshotNames("", datasetName)
	if len(sourceSnapshots) < 1 {
		log.Printf("run error: no source snapshot to replicate %s", datasetName)
		return
	}

	// Only target snapshots of the same frequency as replication
	var filteredList []string
	for _, l := range sourceSnapshots {

		if strings.Contains(l, frequency) {
			filteredList = append(filteredList, l)
		}
	}
	sort.Strings(filteredList)
	sourceSnapshots = filteredList

	for _, d := range to {

		// Find destination snapshots
		destinationHost := ""
		destinationDataset := d
		isRemote := strings.Contains(d, "@")

		var destinationSnapshots []string
		if isRemote {
			destinationDataset = strings.Split(d, "@")[0]
			destinationHost = strings.Split(d, "@")[1]
			destinationSnapshots = zfs.ListSnapshotNames(destinationHost, destinationDataset)
		} else {
			destinationSnapshots = zfs.ListSnapshotNames("", d)
		}
		sort.Strings(destinationSnapshots)

		// Perform new replication
		if len(destinationSnapshots) == 0 {
			latestSnapshot := sourceSnapshots[len(sourceSnapshots)-1]
			err := zfs.Replicate(latestSnapshot, destinationHost, destinationDataset, raw)
			if err != nil {
				log.Printf("run error: failed to replicate %s to %s", datasetName, d)
				log.Printf("run error: %s", err)
				continue
			}

			log.Printf("Ok: replicated %s to %s", datasetName, d)
			continue
		}

		// Perform incremental replication
		latestMatchingSource := -1
		for i, s := range sourceSnapshots {
			for _, d := range destinationSnapshots {

				if strings.Split(d, "@")[1] == strings.Split(s, "@")[1] {
					latestMatchingSource = i
				}
			}
		}

		if latestMatchingSource == -1 {
			log.Printf("run error: no matching snapshot between %s and %s", datasetName, d)
			continue
		}

		toIncrement := sourceSnapshots[latestMatchingSource:]
		for i := range toIncrement {
			if i == len(toIncrement)-1 {
				break
			}

			source1 := toIncrement[i]
			source2 := toIncrement[i+1]

			err := zfs.Increment(source1, source2, destinationHost, destinationDataset, raw)
			if err != nil {
				log.Printf("run error: failed to increment %s to %s", datasetName, d)
				log.Printf("run error: %s", err)
				break
			}

			log.Printf("Ok: incremented %s to %s", source2, d)
		}
	}
}
