package utils

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
)

func GetContainerID() (string, error) {
	cgroup := "/proc/self/cgroup"
	f, err := os.Open(cgroup)
	if err != nil {
		return "", fmt.Errorf("unable to detect cgroup.  are you sure you are in a container? error: %s", err)
	}
	defer f.Close()

	id := ""

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		l := scanner.Text()

		parts := strings.Split(l, ":")
		if len(parts) < 3 {
			continue
		}

		data := parts[2]
		dataParts := strings.Split(data, "/")

		if len(dataParts) < 3 {
			continue
		}

		id = dataParts[2]

		if id != "" {
			return strings.TrimSpace(id), nil
		}
	}

	content, err := ioutil.ReadFile(cgroup)
	if err != nil {
		return "", err
	}
	return "", fmt.Errorf("unable to get container id: %s", string(content))
}
