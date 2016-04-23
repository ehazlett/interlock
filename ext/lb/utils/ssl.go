package utils

import (
	"github.com/ehazlett/interlock/ext"
	"github.com/samalba/dockerclient"
)

const (
	DefaultSSLBackendTLSVerify = "none"
)

func SSLEnabled(config *dockerclient.ContainerConfig) bool {
	if _, ok := config.Labels[ext.InterlockSSLLabel]; ok {
		return true
	}

	return false
}

func SSLOnly(config *dockerclient.ContainerConfig) bool {
	if _, ok := config.Labels[ext.InterlockSSLOnlyLabel]; ok {
		return true
	}

	return false
}

func SSLBackend(config *dockerclient.ContainerConfig) bool {
	if _, ok := config.Labels[ext.InterlockSSLBackendLabel]; ok {
		return true
	}

	return false
}

func SSLCertName(config *dockerclient.ContainerConfig) string {
	if v, ok := config.Labels[ext.InterlockSSLCertLabel]; ok {
		return v
	}

	return ""
}

func SSLCertKey(config *dockerclient.ContainerConfig) string {
	if v, ok := config.Labels[ext.InterlockSSLCertKeyLabel]; ok {
		return v
	}

	return ""
}

func SSLBackendTLSVerify(config *dockerclient.ContainerConfig) string {
	verify := DefaultSSLBackendTLSVerify

	if v, ok := config.Labels[ext.InterlockSSLBackendTLSVerifyLabel]; ok {
		verify = v
	}

	return verify
}
