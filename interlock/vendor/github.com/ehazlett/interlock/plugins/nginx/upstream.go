package nginx

type Server struct {
	Addr string
}

type Upstream struct {
	Name    string
	Servers []*Server
}
