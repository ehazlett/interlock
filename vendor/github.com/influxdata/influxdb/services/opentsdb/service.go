package opentsdb // import "github.com/influxdata/influxdb/services/opentsdb"

import (
	"bufio"
	"bytes"
	"crypto/tls"
<<<<<<< HEAD
	"expvar"
=======
>>>>>>> 12a5469... start on swarm services; move to glade
	"io"
	"log"
	"net"
	"net/http"
	"net/textproto"
	"os"
	"strconv"
	"strings"
	"sync"
<<<<<<< HEAD
	"time"

	"github.com/influxdata/influxdb"
=======
	"sync/atomic"
	"time"

>>>>>>> 12a5469... start on swarm services; move to glade
	"github.com/influxdata/influxdb/models"
	"github.com/influxdata/influxdb/services/meta"
	"github.com/influxdata/influxdb/tsdb"
)

// statistics gathered by the openTSDB package.
const (
	statHTTPConnectionsHandled   = "httpConnsHandled"
	statTelnetConnectionsActive  = "tlConnsActive"
	statTelnetConnectionsHandled = "tlConnsHandled"
	statTelnetPointsReceived     = "tlPointsRx"
	statTelnetBytesReceived      = "tlBytesRx"
	statTelnetReadError          = "tlReadErr"
	statTelnetBadLine            = "tlBadLine"
	statTelnetBadTime            = "tlBadTime"
	statTelnetBadTag             = "tlBadTag"
	statTelnetBadFloat           = "tlBadFloat"
<<<<<<< HEAD
	statBatchesTrasmitted        = "batchesTx"
=======
	statBatchesTransmitted       = "batchesTx"
>>>>>>> 12a5469... start on swarm services; move to glade
	statPointsTransmitted        = "pointsTx"
	statBatchesTransmitFail      = "batchesTxFail"
	statConnectionsActive        = "connsActive"
	statConnectionsHandled       = "connsHandled"
	statDroppedPointsInvalid     = "droppedPointsInvalid"
)

// Service manages the listener and handler for an HTTP endpoint.
type Service struct {
	ln     net.Listener  // main listener
	httpln *chanListener // http channel-based listener

	mu   sync.Mutex
	wg   sync.WaitGroup
	done chan struct{}
	err  chan error
	tls  bool
	cert string

	BindAddress     string
	Database        string
	RetentionPolicy string

	PointsWriter interface {
		WritePoints(database, retentionPolicy string, consistencyLevel models.ConsistencyLevel, points []models.Point) error
	}
	MetaClient interface {
		CreateDatabase(name string) (*meta.DatabaseInfo, error)
	}

	// Points received over the telnet protocol are batched.
	batchSize    int
	batchPending int
	batchTimeout time.Duration
	batcher      *tsdb.PointBatcher

	LogPointErrors bool
	Logger         *log.Logger
<<<<<<< HEAD
	statMap        *expvar.Map
=======

	stats    *Statistics
	statTags models.Tags
>>>>>>> 12a5469... start on swarm services; move to glade
}

// NewService returns a new instance of Service.
func NewService(c Config) (*Service, error) {
	// Use defaults where necessary.
	d := c.WithDefaults()

	s := &Service{
		done:            make(chan struct{}),
		tls:             d.TLSEnabled,
		cert:            d.Certificate,
		err:             make(chan error),
		BindAddress:     d.BindAddress,
		Database:        d.Database,
		RetentionPolicy: d.RetentionPolicy,
		batchSize:       d.BatchSize,
		batchPending:    d.BatchPending,
		batchTimeout:    time.Duration(d.BatchTimeout),
		Logger:          log.New(os.Stderr, "[opentsdb] ", log.LstdFlags),
		LogPointErrors:  d.LogPointErrors,
<<<<<<< HEAD
=======
		stats:           &Statistics{},
		statTags:        map[string]string{"bind": d.BindAddress},
>>>>>>> 12a5469... start on swarm services; move to glade
	}
	return s, nil
}

// Open starts the service
func (s *Service) Open() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.Logger.Println("Starting OpenTSDB service")

<<<<<<< HEAD
	// Configure expvar monitoring. It's OK to do this even if the service fails to open and
	// should be done before any data could arrive for the service.
	key := strings.Join([]string{"opentsdb", s.BindAddress}, ":")
	tags := map[string]string{"bind": s.BindAddress}
	s.statMap = influxdb.NewStatistics(key, "opentsdb", tags)

=======
>>>>>>> 12a5469... start on swarm services; move to glade
	if _, err := s.MetaClient.CreateDatabase(s.Database); err != nil {
		s.Logger.Printf("Failed to ensure target database %s exists: %s", s.Database, err.Error())
		return err
	}

	s.batcher = tsdb.NewPointBatcher(s.batchSize, s.batchPending, s.batchTimeout)
	s.batcher.Start()

	// Start processing batches.
	s.wg.Add(1)
	go s.processBatches(s.batcher)

	// Open listener.
	if s.tls {
		cert, err := tls.LoadX509KeyPair(s.cert, s.cert)
		if err != nil {
			return err
		}

		listener, err := tls.Listen("tcp", s.BindAddress, &tls.Config{
			Certificates: []tls.Certificate{cert},
		})
		if err != nil {
			return err
		}

		s.Logger.Println("Listening on TLS:", listener.Addr().String())
		s.ln = listener
	} else {
		listener, err := net.Listen("tcp", s.BindAddress)
		if err != nil {
			return err
		}

		s.Logger.Println("Listening on:", listener.Addr().String())
		s.ln = listener
	}
	s.httpln = newChanListener(s.ln.Addr())

	// Begin listening for connections.
	s.wg.Add(2)
	go s.serveHTTP()
	go s.serve()

	return nil
}

// Close closes the openTSDB service
func (s *Service) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.ln != nil {
		return s.ln.Close()
	}

	if s.batcher != nil {
		s.batcher.Stop()
	}
	close(s.done)
	s.wg.Wait()
	return nil
}

// SetLogOutput sets the writer to which all logs are written. It must not be
// called after Open is called.
func (s *Service) SetLogOutput(w io.Writer) {
	s.Logger = log.New(w, "[opentsdb] ", log.LstdFlags)
}

<<<<<<< HEAD
=======
// Statistics maintains statistics for the subscriber service.
type Statistics struct {
	HTTPConnectionsHandled   int64
	ActiveTelnetConnections  int64
	HandledTelnetConnections int64
	TelnetPointsReceived     int64
	TelnetBytesReceived      int64
	TelnetReadError          int64
	TelnetBadLine            int64
	TelnetBadTime            int64
	TelnetBadTag             int64
	TelnetBadFloat           int64
	BatchesTransmitted       int64
	PointsTransmitted        int64
	BatchesTransmitFail      int64
	ActiveConnections        int64
	HandledConnections       int64
	InvalidDroppedPoints     int64
}

// Statistics returns statistics for periodic monitoring.
func (s *Service) Statistics(tags map[string]string) []models.Statistic {
	return []models.Statistic{{
		Name: "opentsdb",
		Tags: s.statTags,
		Values: map[string]interface{}{
			statHTTPConnectionsHandled:   atomic.LoadInt64(&s.stats.HTTPConnectionsHandled),
			statTelnetConnectionsActive:  atomic.LoadInt64(&s.stats.ActiveTelnetConnections),
			statTelnetConnectionsHandled: atomic.LoadInt64(&s.stats.HandledTelnetConnections),
			statTelnetPointsReceived:     atomic.LoadInt64(&s.stats.TelnetPointsReceived),
			statTelnetBytesReceived:      atomic.LoadInt64(&s.stats.TelnetBytesReceived),
			statTelnetReadError:          atomic.LoadInt64(&s.stats.TelnetReadError),
			statTelnetBadLine:            atomic.LoadInt64(&s.stats.TelnetBadLine),
			statTelnetBadTime:            atomic.LoadInt64(&s.stats.TelnetBadTime),
			statTelnetBadTag:             atomic.LoadInt64(&s.stats.TelnetBadTag),
			statTelnetBadFloat:           atomic.LoadInt64(&s.stats.TelnetBadFloat),
			statBatchesTransmitted:       atomic.LoadInt64(&s.stats.BatchesTransmitted),
			statPointsTransmitted:        atomic.LoadInt64(&s.stats.PointsTransmitted),
			statBatchesTransmitFail:      atomic.LoadInt64(&s.stats.BatchesTransmitFail),
			statConnectionsActive:        atomic.LoadInt64(&s.stats.ActiveConnections),
			statConnectionsHandled:       atomic.LoadInt64(&s.stats.HandledConnections),
			statDroppedPointsInvalid:     atomic.LoadInt64(&s.stats.InvalidDroppedPoints),
		},
	}}
}

>>>>>>> 12a5469... start on swarm services; move to glade
// Err returns a channel for fatal errors that occur on the listener.
func (s *Service) Err() <-chan error { return s.err }

// Addr returns the listener's address. Returns nil if listener is closed.
func (s *Service) Addr() net.Addr {
	if s.ln == nil {
		return nil
	}
	return s.ln.Addr()
}

// serve serves the handler from the listener.
func (s *Service) serve() {
	defer s.wg.Done()

	for {
		// Wait for next connection.
		conn, err := s.ln.Accept()
		if opErr, ok := err.(*net.OpError); ok && !opErr.Temporary() {
			s.Logger.Println("openTSDB TCP listener closed")
			return
		} else if err != nil {
			s.Logger.Println("error accepting openTSDB: ", err.Error())
			continue
		}

		// Handle connection in separate goroutine.
		go s.handleConn(conn)
	}
}

// handleConn processes conn. This is run in a separate goroutine.
func (s *Service) handleConn(conn net.Conn) {
<<<<<<< HEAD
	defer s.statMap.Add(statConnectionsActive, -1)
	s.statMap.Add(statConnectionsActive, 1)
	s.statMap.Add(statConnectionsHandled, 1)
=======
	defer atomic.AddInt64(&s.stats.ActiveConnections, -1)
	atomic.AddInt64(&s.stats.ActiveConnections, 1)
	atomic.AddInt64(&s.stats.HandledConnections, 1)
>>>>>>> 12a5469... start on swarm services; move to glade

	// Read header into buffer to check if it's HTTP.
	var buf bytes.Buffer
	r := bufio.NewReader(io.TeeReader(conn, &buf))

	// Attempt to parse connection as HTTP.
	_, err := http.ReadRequest(r)

	// Rebuild connection from buffer and remaining connection data.
	bufr := bufio.NewReader(io.MultiReader(&buf, conn))
	conn = &readerConn{Conn: conn, r: bufr}

	// If no HTTP parsing error occurred then process as HTTP.
	if err == nil {
<<<<<<< HEAD
		s.statMap.Add(statHTTPConnectionsHandled, 1)
=======
		atomic.AddInt64(&s.stats.HTTPConnectionsHandled, 1)
>>>>>>> 12a5469... start on swarm services; move to glade
		s.httpln.ch <- conn
		return
	}

	// Otherwise handle in telnet format.
	s.wg.Add(1)
	s.handleTelnetConn(conn)
}

// handleTelnetConn accepts OpenTSDB's telnet protocol.
// Each telnet command consists of a line of the form:
//   put sys.cpu.user 1356998400 42.5 host=webserver01 cpu=0
func (s *Service) handleTelnetConn(conn net.Conn) {
	defer conn.Close()
	defer s.wg.Done()
<<<<<<< HEAD
	defer s.statMap.Add(statTelnetConnectionsActive, -1)
	s.statMap.Add(statTelnetConnectionsActive, 1)
	s.statMap.Add(statTelnetConnectionsHandled, 1)
=======
	defer atomic.AddInt64(&s.stats.ActiveTelnetConnections, -1)
	atomic.AddInt64(&s.stats.ActiveTelnetConnections, 1)
	atomic.AddInt64(&s.stats.HandledTelnetConnections, 1)
>>>>>>> 12a5469... start on swarm services; move to glade

	// Get connection details.
	remoteAddr := conn.RemoteAddr().String()

	// Wrap connection in a text protocol reader.
	r := textproto.NewReader(bufio.NewReader(conn))
	for {
		line, err := r.ReadLine()
		if err != nil {
			if err != io.EOF {
<<<<<<< HEAD
				s.statMap.Add(statTelnetReadError, 1)
=======
				atomic.AddInt64(&s.stats.TelnetReadError, 1)
>>>>>>> 12a5469... start on swarm services; move to glade
				s.Logger.Println("error reading from openTSDB connection", err.Error())
			}
			return
		}
<<<<<<< HEAD
		s.statMap.Add(statTelnetPointsReceived, 1)
		s.statMap.Add(statTelnetBytesReceived, int64(len(line)))
=======
		atomic.AddInt64(&s.stats.TelnetPointsReceived, 1)
		atomic.AddInt64(&s.stats.TelnetBytesReceived, int64(len(line)))
>>>>>>> 12a5469... start on swarm services; move to glade

		inputStrs := strings.Fields(line)

		if len(inputStrs) == 1 && inputStrs[0] == "version" {
			conn.Write([]byte("InfluxDB TSDB proxy"))
			continue
		}

		if len(inputStrs) < 4 || inputStrs[0] != "put" {
<<<<<<< HEAD
			s.statMap.Add(statTelnetBadLine, 1)
=======
			atomic.AddInt64(&s.stats.TelnetBadLine, 1)
>>>>>>> 12a5469... start on swarm services; move to glade
			if s.LogPointErrors {
				s.Logger.Printf("malformed line '%s' from %s", line, remoteAddr)
			}
			continue
		}

		measurement := inputStrs[1]
		tsStr := inputStrs[2]
		valueStr := inputStrs[3]
		tagStrs := inputStrs[4:]

		var t time.Time
		ts, err := strconv.ParseInt(tsStr, 10, 64)
		if err != nil {
<<<<<<< HEAD
			s.statMap.Add(statTelnetBadTime, 1)
=======
			atomic.AddInt64(&s.stats.TelnetBadTime, 1)
>>>>>>> 12a5469... start on swarm services; move to glade
			if s.LogPointErrors {
				s.Logger.Printf("malformed time '%s' from %s", tsStr, remoteAddr)
			}
		}

		switch len(tsStr) {
		case 10:
			t = time.Unix(ts, 0)
			break
		case 13:
			t = time.Unix(ts/1000, (ts%1000)*1000)
			break
		default:
<<<<<<< HEAD
			s.statMap.Add(statTelnetBadTime, 1)
=======
			atomic.AddInt64(&s.stats.TelnetBadTime, 1)
>>>>>>> 12a5469... start on swarm services; move to glade
			if s.LogPointErrors {
				s.Logger.Printf("bad time '%s' must be 10 or 13 chars, from %s ", tsStr, remoteAddr)
			}
			continue
		}

		tags := make(map[string]string)
		for t := range tagStrs {
			parts := strings.SplitN(tagStrs[t], "=", 2)
			if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
<<<<<<< HEAD
				s.statMap.Add(statTelnetBadTag, 1)
=======
				atomic.AddInt64(&s.stats.TelnetBadTag, 1)
>>>>>>> 12a5469... start on swarm services; move to glade
				if s.LogPointErrors {
					s.Logger.Printf("malformed tag data '%v' from %s", tagStrs[t], remoteAddr)
				}
				continue
			}
			k := parts[0]

			tags[k] = parts[1]
		}

		fields := make(map[string]interface{})
		fv, err := strconv.ParseFloat(valueStr, 64)
		if err != nil {
<<<<<<< HEAD
			s.statMap.Add(statTelnetBadFloat, 1)
=======
			atomic.AddInt64(&s.stats.TelnetBadFloat, 1)
>>>>>>> 12a5469... start on swarm services; move to glade
			if s.LogPointErrors {
				s.Logger.Printf("bad float '%s' from %s", valueStr, remoteAddr)
			}
			continue
		}
		fields["value"] = fv

		pt, err := models.NewPoint(measurement, tags, fields, t)
		if err != nil {
<<<<<<< HEAD
			s.statMap.Add(statTelnetBadFloat, 1)
=======
			atomic.AddInt64(&s.stats.TelnetBadFloat, 1)
>>>>>>> 12a5469... start on swarm services; move to glade
			if s.LogPointErrors {
				s.Logger.Printf("bad float '%s' from %s", valueStr, remoteAddr)
			}
			continue
		}
		s.batcher.In() <- pt
	}
}

// serveHTTP handles connections in HTTP format.
func (s *Service) serveHTTP() {
<<<<<<< HEAD
	srv := &http.Server{Handler: &Handler{
=======
	handler := &Handler{
>>>>>>> 12a5469... start on swarm services; move to glade
		Database:        s.Database,
		RetentionPolicy: s.RetentionPolicy,
		PointsWriter:    s.PointsWriter,
		Logger:          s.Logger,
<<<<<<< HEAD
		statMap:         s.statMap,
	}}
=======
		stats:           s.stats,
	}
	srv := &http.Server{Handler: handler}
>>>>>>> 12a5469... start on swarm services; move to glade
	srv.Serve(s.httpln)
}

// processBatches continually drains the given batcher and writes the batches to the database.
func (s *Service) processBatches(batcher *tsdb.PointBatcher) {
	defer s.wg.Done()
	for {
		select {
		case batch := <-batcher.Out():
			if err := s.PointsWriter.WritePoints(s.Database, s.RetentionPolicy, models.ConsistencyLevelAny, batch); err == nil {
<<<<<<< HEAD
				s.statMap.Add(statBatchesTrasmitted, 1)
				s.statMap.Add(statPointsTransmitted, int64(len(batch)))
			} else {
				s.Logger.Printf("failed to write point batch to database %q: %s", s.Database, err)
				s.statMap.Add(statBatchesTransmitFail, 1)
=======
				atomic.AddInt64(&s.stats.BatchesTransmitted, 1)
				atomic.AddInt64(&s.stats.PointsTransmitted, int64(len(batch)))
			} else {
				s.Logger.Printf("failed to write point batch to database %q: %s", s.Database, err)
				atomic.AddInt64(&s.stats.BatchesTransmitFail, 1)
>>>>>>> 12a5469... start on swarm services; move to glade
			}

		case <-s.done:
			return
		}
	}
}
