package lb

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

func getContainerID() (string, error) {
	f, err := os.Open("/proc/self/cgroup")
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

	return "", fmt.Errorf("unable to get container id")
}
