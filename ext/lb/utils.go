package lb

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

func getNodeID() (string, error) {
	f, err := os.Open("/proc/self/cgroup")
	if err != nil {
		return "", fmt.Errorf("unable to detect cgroup.  are you sure you are in a container? error: %s", err)
	}
	defer f.Close()

	r := bufio.NewReader(f)
	l, err := r.ReadString('\n')
	if err != nil {
		return "", err
	}

	parts := strings.Split(l, ":")
	if len(parts) < 2 {
		return "", fmt.Errorf("unable to parse cgroups for ID")
	}

	data := parts[2]
	dataParts := strings.Split(data, "/")

	if len(dataParts) < 2 {
		return "", fmt.Errorf("unable to parse ID")
	}

	id := dataParts[2]

<<<<<<< HEAD
	return strings.TrimSpace(id), nil
=======
	return id, nil
>>>>>>> e2cf8e1... calculate proxy container restart among interlock instances
}
