package zfs

import (
	"fmt"
	"os/exec"
	"sort"
	"strings"
)

func Exists(host, zfsType, name string) bool {

	var cmd *exec.Cmd
	if host != "" {
		if zfsType == "pool" {
			cmd = exec.Command("ssh", host, "zpool", "list", "-H", "-o", "name", name)
		} else {
			cmd = exec.Command("ssh", host, "zfs", "list", "-H", "-o", "name", name)
		}

	} else {
		if zfsType == "pool" {
			cmd = exec.Command("zpool", "list", "-H", "-o", "name", name)
		} else {
			cmd = exec.Command("zfs", "list", "-H", "-o", "name", name)
		}

	}

	output, err := cmd.CombinedOutput()
	if err != nil {
		return false
	}

	if name != strings.TrimSpace(string(output)) {
		return false
	}

	return true
}

func ListSnapshotNames(host, name string) []string {

	cmd := exec.Command("zfs", "list", "-H", "-t", "snapshot", "-o", "name", name)
	if host != "" {
		cmd = exec.Command("ssh", host, "zfs", "list", "-H", "-t", "snapshot", "-o", "name", name)
	}

	output, err := cmd.CombinedOutput()
	if err != nil {
		return []string{}
	}

	out := string(output)
	list := strings.Split(out, "\n")
	return list[:len(list)-1]
}

func ListChildsNamesOf(name string) []string {

	cmd := exec.Command("zfs", "list", "-H", "-t", "filesystem,volume", "-o", "name")

	output, err := cmd.CombinedOutput()
	if err != nil {
		return []string{}
	}

	out := string(output)
	list := strings.Split(out, "\n")

	result := []string{}
	for _, l := range list {
		if l == name {
			continue
		}

		if strings.Contains(l, name) {
			result = append(result, l)
		}
	}

	return result
}

func Snapshot(datasetName, snapshot_name string) error {

	cmd := exec.Command("zfs", "snapshot", datasetName+"@"+snapshot_name)
	output, err := cmd.CombinedOutput()
	out := string(output)
	if err != nil {
		return fmt.Errorf("snapshot error: %s %s (%s)", err, out, datasetName)
	}

	return nil
}

func Prune(datasetName, suffix string, retention int) ([]string, error) {

	cmd := exec.Command("zfs", "list", "-H", "-t", "snapshot", "-o", "name", datasetName)
	output, err := cmd.CombinedOutput()
	out := string(output)
	if err != nil {
		return []string{}, fmt.Errorf("prune error: %s %s (%s)", err, out, datasetName)
	}

	list := strings.Split(out, "\n")

	var filteredList []string
	for _, l := range list {

		if strings.Contains(l, suffix) {
			filteredList = append(filteredList, l)
		}
	}
	sort.Strings(filteredList)

	if len(filteredList) <= retention {
		return []string{}, nil
	}

	filteredList = filteredList[:len(filteredList)-retention]
	for _, l := range filteredList {
		err = Destroy(l)
		if err != nil {
			return []string{}, fmt.Errorf("prune error: %s (%s)", err, l)
		}
	}

	return filteredList, nil
}

func Scrub(poolName string) error {

	cmd := exec.Command("zpool", "scrub", poolName)
	output, err := cmd.CombinedOutput()
	out := string(output)
	if err != nil {
		return fmt.Errorf("scrub error: %s %s (%s)", err, out, poolName)
	}

	return nil
}

func Destroy(datasetName string) error {

	cmd := exec.Command("zfs", "destroy", datasetName)
	output, err := cmd.CombinedOutput()
	out := string(output)
	if err != nil {
		return fmt.Errorf("destroy error: %s %s (%s)", err, out, datasetName)
	}

	return nil
}

func Replicate(source, destinationHost, destinationDataset string, raw bool) error {

	sshCmd := ""
	if destinationHost != "" {
		sshCmd = fmt.Sprintf("ssh %s", destinationHost)
	}
	rawCmd := ""
	if raw {
		rawCmd = "--raw"
	}

	replicateCmd := fmt.Sprintf("zfs send %s --large-block %s | %s zfs receive -F %s", rawCmd, source, sshCmd, destinationDataset)
	cmd := exec.Command("sh", "-c", replicateCmd)
	stdout, err := cmd.CombinedOutput()
	out := string(stdout)
	if err != nil {
		return fmt.Errorf("replicate error: %s %s (%s to %s)", err, out, source, destinationDataset)
	}

	return nil
}

func Increment(source1, source2, destinationHost, destinationDataset string, raw bool) error {

	sshCmd := ""
	if destinationHost != "" {
		sshCmd = fmt.Sprintf("ssh %s", destinationHost)
	}
	rawCmd := ""
	if raw {
		rawCmd = "--raw"
	}

	replicateCmd := fmt.Sprintf("zfs send %s --large-block -i %s %s | %s zfs receive -F %s", rawCmd, source1, source2, sshCmd, destinationDataset)
	cmd := exec.Command("sh", "-c", replicateCmd)
	stdout, err := cmd.CombinedOutput()
	out := string(stdout)
	if err != nil {
		return fmt.Errorf("increment error: %s %s (diff from %s & %s to %s)", err, out, source1, source2, destinationDataset)
	}

	return nil
}
