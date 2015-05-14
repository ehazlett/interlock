package nginx

type Host struct {
	ServerNames []string
	ListenPort  int
	Upstream    *Upstream
}
