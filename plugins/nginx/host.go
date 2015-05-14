package nginx

type Host struct {
	ServerNames []string
	Port        int
	SSLPort     int
	SSL         bool
	SSLCert     string
	SSLCertKey  string
	SSLOnly     bool
	Upstream    *Upstream
}
