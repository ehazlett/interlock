package haproxy

import (
	"fmt"
	"os/exec"
	"time"
)

func (p *HAProxyLoadBalancer) configIPTables(drop bool) error {
	ports := []int{
		p.cfg.Port,
	}

	if p.cfg.SSLPort != 0 {
		ports = append(ports, p.cfg.SSLPort)
	}

	d := "-I"

	if drop {
		d = "-D"
	}

	iptables, err := exec.LookPath("iptables")
	if err != nil {
		return err
	}

	for _, port := range ports {
		args := []string{
			d,
			"INPUT",
			"-p",
			"tcp",
			"--dport",
			fmt.Sprintf("%d", port),
			"--syn",
			"-j",
			"DROP",
		}

		cmd := exec.Command(iptables, args...)

		log().Debug(cmd)

		out, err := cmd.Output()
		if err != nil {
			return err
		}

		log().Debugf("iptables out: %s", string(out))
	}

	return nil
}

func (p *HAProxyLoadBalancer) dropSYN() error {
	log().Debug("dropping SYN packets to trigger client re-send")

	if err := p.configIPTables(false); err != nil {
		return err
	}

	// pause to make clients drop
	time.Sleep(time.Second * 1)

	return nil
}

func (p *HAProxyLoadBalancer) resumeSYN() error {
	log().Debug("resuming SYN packets")

	if err := p.configIPTables(false); err != nil {
		return err
	}

	return nil
}
