package httpd

<<<<<<< HEAD
// DefaultBindAddress is the default address to bind to.
const DefaultBindAddress = ":8086"
=======
const (
	// DefaultBindAddress is the default address to bind to.
	DefaultBindAddress = ":8086"

	// DefaultRealm is the default realm sent back when issuing a basic auth challenge.
	DefaultRealm = "InfluxDB"
)
>>>>>>> 12a5469... start on swarm services; move to glade

// Config represents a configuration for a HTTP service.
type Config struct {
	Enabled            bool   `toml:"enabled"`
	BindAddress        string `toml:"bind-address"`
	AuthEnabled        bool   `toml:"auth-enabled"`
	LogEnabled         bool   `toml:"log-enabled"`
	WriteTracing       bool   `toml:"write-tracing"`
	HTTPSEnabled       bool   `toml:"https-enabled"`
	HTTPSCertificate   string `toml:"https-certificate"`
	HTTPSPrivateKey    string `toml:"https-private-key"`
	MaxRowLimit        int    `toml:"max-row-limit"`
	MaxConnectionLimit int    `toml:"max-connection-limit"`
	SharedSecret       string `toml:"shared-secret"`
<<<<<<< HEAD
=======
	Realm              string `toml:"realm"`
>>>>>>> 12a5469... start on swarm services; move to glade
}

// NewConfig returns a new Config with default settings.
func NewConfig() Config {
	return Config{
		Enabled:          true,
<<<<<<< HEAD
		BindAddress:      ":8086",
=======
		BindAddress:      DefaultBindAddress,
>>>>>>> 12a5469... start on swarm services; move to glade
		LogEnabled:       true,
		HTTPSEnabled:     false,
		HTTPSCertificate: "/etc/ssl/influxdb.pem",
		MaxRowLimit:      DefaultChunkSize,
<<<<<<< HEAD
=======
		Realm:            DefaultRealm,
>>>>>>> 12a5469... start on swarm services; move to glade
	}
}
