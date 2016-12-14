package utils

import "github.com/ehazlett/interlock/ext"

const (
	DefaultSSLBackendTLSVerify = "none"
)

func SSLEnabled(labels map[string]string) bool {
	if _, ok := labels[ext.InterlockSSLLabel]; ok {
		return true
	}

	return false
}

func SSLOnly(labels map[string]string) bool {
	if _, ok := labels[ext.InterlockSSLOnlyLabel]; ok {
		return true
	}

	return false
}

func SSLBackend(labels map[string]string) bool {
	if _, ok := labels[ext.InterlockSSLBackendLabel]; ok {
		return true
	}

	return false
}

func SSLCertName(labels map[string]string) string {
	if v, ok := labels[ext.InterlockSSLCertLabel]; ok {
		return v
	}

	return ""
}

func SSLCertKey(labels map[string]string) string {
	if v, ok := labels[ext.InterlockSSLCertKeyLabel]; ok {
		return v
	}

	return ""
}

func SSLBackendTLSVerify(labels map[string]string) string {
	verify := DefaultSSLBackendTLSVerify

	if v, ok := labels[ext.InterlockSSLBackendTLSVerifyLabel]; ok {
		verify = v
	}

	return verify
}
