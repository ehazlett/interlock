package portmapper

import "net"

<<<<<<< HEAD
func newMockProxyCommand(proto string, hostIP net.IP, hostPort int, containerIP net.IP, containerPort int) userlandProxy {
	return &mockProxyCommand{}
=======
func newMockProxyCommand(proto string, hostIP net.IP, hostPort int, containerIP net.IP, containerPort int) (userlandProxy, error) {
	return &mockProxyCommand{}, nil
>>>>>>> 12a5469... start on swarm services; move to glade
}

type mockProxyCommand struct {
}

func (p *mockProxyCommand) Start() error {
	return nil
}

func (p *mockProxyCommand) Stop() error {
	return nil
}
