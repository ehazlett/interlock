package utils

import (
	"github.com/docker/docker/api/types"
	"github.com/ehazlett/interlock/ext"
)

const (
	DefaultSSLBackendTLSVerify = "none"
)

func SSLEnabled(config types.Container) bool {
	if _, ok := config.Labels[ext.InterlockSSLLabel]; ok {
		return true
	}

	return false
}

func SSLOnly(config types.Container) bool {
	if _, ok := config.Labels[ext.InterlockSSLOnlyLabel]; ok {
		return true
	}

	return false
}

func SSLBackend(config types.Container) bool {
	if _, ok := config.Labels[ext.InterlockSSLBackendLabel]; ok {
		return true
	}

	return false
}

func SSLCertName(config types.Container) string {
	if v, ok := config.Labels[ext.InterlockSSLCertLabel]; ok {
		return v
	}

	return ""
}

func SSLCertKey(config types.Container) string {
	if v, ok := config.Labels[ext.InterlockSSLCertKeyLabel]; ok {
		return v
	}

	return ""
}

func SSLBackendTLSVerify(config types.Container) string {
	verify := DefaultSSLBackendTLSVerify

	if v, ok := config.Labels[ext.InterlockSSLBackendTLSVerifyLabel]; ok {
		verify = v
	}

	return verify
}
