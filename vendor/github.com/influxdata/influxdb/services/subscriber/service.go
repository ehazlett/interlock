package subscriber // import "github.com/influxdata/influxdb/services/subscriber"

import (
	"errors"
<<<<<<< HEAD
	"expvar"
=======
>>>>>>> 12a5469... start on swarm services; move to glade
	"fmt"
	"io"
	"log"
	"net/url"
	"os"
<<<<<<< HEAD
	"strings"
	"sync"
	"time"

	"github.com/influxdata/influxdb"
	"github.com/influxdata/influxdb/coordinator"
=======
	"sync"
	"sync/atomic"
	"time"

	"github.com/influxdata/influxdb/coordinator"
	"github.com/influxdata/influxdb/models"
	"github.com/influxdata/influxdb/monitor"
>>>>>>> 12a5469... start on swarm services; move to glade
	"github.com/influxdata/influxdb/services/meta"
)

// Statistics for the Subscriber service.
const (
	statPointsWritten = "pointsWritten"
	statWriteFailures = "writeFailures"
)

// PointsWriter is an interface for writing points to a subscription destination.
// Only WritePoints() needs to be satisfied.
type PointsWriter interface {
	WritePoints(p *coordinator.WritePointsRequest) error
}

// unique set that identifies a given subscription
type subEntry struct {
	db   string
	rp   string
	name string
}

// Service manages forking the incoming data from InfluxDB
// to defined third party destinations.
// Subscriptions are defined per database and retention policy.
type Service struct {
	MetaClient interface {
		Databases() []meta.DatabaseInfo
		WaitForDataChanged() chan struct{}
	}
	NewPointsWriter func(u url.URL) (PointsWriter, error)
	Logger          *log.Logger
	update          chan struct{}
<<<<<<< HEAD
	statMap         *expvar.Map
=======
	stats           *Statistics
>>>>>>> 12a5469... start on swarm services; move to glade
	points          chan *coordinator.WritePointsRequest
	wg              sync.WaitGroup
	closed          bool
	closing         chan struct{}
	mu              sync.Mutex
	conf            Config

<<<<<<< HEAD
	failures      *expvar.Int
	pointsWritten *expvar.Int
=======
	subs  map[subEntry]chanWriter
	subMu sync.RWMutex
>>>>>>> 12a5469... start on swarm services; move to glade
}

// NewService returns a subscriber service with given settings
func NewService(c Config) *Service {
	s := &Service{
<<<<<<< HEAD
		Logger:        log.New(os.Stderr, "[subscriber] ", log.LstdFlags),
		statMap:       influxdb.NewStatistics("subscriber", "subscriber", nil),
		closed:        true,
		conf:          c,
		failures:      &expvar.Int{},
		pointsWritten: &expvar.Int{},
	}
	s.NewPointsWriter = s.newPointsWriter
	s.statMap.Set(statWriteFailures, s.failures)
	s.statMap.Set(statPointsWritten, s.pointsWritten)
=======
		Logger: log.New(os.Stderr, "[subscriber] ", log.LstdFlags),
		closed: true,
		stats:  &Statistics{},
		conf:   c,
	}
	s.NewPointsWriter = s.newPointsWriter
>>>>>>> 12a5469... start on swarm services; move to glade
	return s
}

// Open starts the subscription service.
func (s *Service) Open() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.MetaClient == nil {
		return errors.New("no meta store")
	}

	s.closed = false

	s.closing = make(chan struct{})
	s.update = make(chan struct{})
	s.points = make(chan *coordinator.WritePointsRequest, 100)

	s.wg.Add(2)
	go func() {
		defer s.wg.Done()
		s.run()
	}()
	go func() {
		defer s.wg.Done()
		s.waitForMetaUpdates()
	}()

	s.Logger.Println("opened service")
	return nil
}

// Close terminates the subscription service
// Will panic if called multiple times or without first opening the service.
func (s *Service) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.closed = true

	close(s.points)
	close(s.closing)

	s.wg.Wait()
	s.Logger.Println("closed service")
	return nil
}

// SetLogOutput sets the writer to which all logs are written. It must not be
// called after Open is called.
func (s *Service) SetLogOutput(w io.Writer) {
	s.Logger = log.New(w, "[subscriber] ", log.LstdFlags)
}

<<<<<<< HEAD
=======
// Statistics maintains the statistics for the subscriber service.
type Statistics struct {
	WriteFailures int64
	PointsWritten int64
}

// Statistics returns statistics for periodic monitoring.
func (s *Service) Statistics(tags map[string]string) []models.Statistic {
	statistics := []models.Statistic{{
		Name: "subscriber",
		Tags: tags,
		Values: map[string]interface{}{
			statPointsWritten: atomic.LoadInt64(&s.stats.PointsWritten),
			statWriteFailures: atomic.LoadInt64(&s.stats.WriteFailures),
		},
	}}

	s.subMu.RLock()
	defer s.subMu.RUnlock()

	for _, sub := range s.subs {
		statistics = append(statistics, sub.Statistics(tags)...)
	}
	return statistics
}

>>>>>>> 12a5469... start on swarm services; move to glade
func (s *Service) waitForMetaUpdates() {
	for {
		ch := s.MetaClient.WaitForDataChanged()
		select {
		case <-ch:
			err := s.Update()
			if err != nil {
				s.Logger.Println("error updating subscriptions:", err)
			}
		case <-s.closing:
			return
		}
	}
}

// Update will start new and stop deleted subscriptions.
func (s *Service) Update() error {
	// signal update
	select {
	case s.update <- struct{}{}:
		return nil
	case <-s.closing:
		return errors.New("service closed cannot update")
	}
}

func (s *Service) createSubscription(se subEntry, mode string, destinations []string) (PointsWriter, error) {
	var bm BalanceMode
	switch mode {
	case "ALL":
		bm = ALL
	case "ANY":
		bm = ANY
	default:
		return nil, fmt.Errorf("unknown balance mode %q", mode)
	}
	writers := make([]PointsWriter, len(destinations))
<<<<<<< HEAD
	statMaps := make([]*expvar.Map, len(writers))
=======
	stats := make([]writerStats, len(writers))
>>>>>>> 12a5469... start on swarm services; move to glade
	for i, dest := range destinations {
		u, err := url.Parse(dest)
		if err != nil {
			return nil, err
		}
		w, err := s.NewPointsWriter(*u)
		if err != nil {
			return nil, err
		}
		writers[i] = w
<<<<<<< HEAD
		tags := map[string]string{
=======
		stats[i].dest = dest
	}
	return &balancewriter{
		bm:      bm,
		writers: writers,
		stats:   stats,
		tags: map[string]string{
>>>>>>> 12a5469... start on swarm services; move to glade
			"database":         se.db,
			"retention_policy": se.rp,
			"name":             se.name,
			"mode":             mode,
<<<<<<< HEAD
			"destination":      dest,
		}
		key := strings.Join([]string{"subscriber", se.db, se.rp, se.name, dest}, ":")
		statMaps[i] = influxdb.NewStatistics(key, "subscriber", tags)
	}
	return &balancewriter{
		bm:       bm,
		writers:  writers,
		statMaps: statMaps,
=======
		},
>>>>>>> 12a5469... start on swarm services; move to glade
	}, nil
}

// Points returns a channel into which write point requests can be sent.
func (s *Service) Points() chan<- *coordinator.WritePointsRequest {
	return s.points
}

// read points off chan and write them
func (s *Service) run() {
	var wg sync.WaitGroup
<<<<<<< HEAD
	subs := make(map[subEntry]chanWriter)
	// Perform initial update
	s.updateSubs(subs, &wg)
	for {
		select {
		case <-s.update:
			err := s.updateSubs(subs, &wg)
=======
	s.subs = make(map[subEntry]chanWriter)
	// Perform initial update
	s.updateSubs(&wg)
	for {
		select {
		case <-s.update:
			err := s.updateSubs(&wg)
>>>>>>> 12a5469... start on swarm services; move to glade
			if err != nil {
				s.Logger.Println("failed to update subscriptions:", err)
			}
		case p, ok := <-s.points:
			if !ok {
				// Close out all chanWriters
<<<<<<< HEAD
				for _, cw := range subs {
					cw.Close()
				}
				// Wait for them to finish
				wg.Wait()
				return
			}
			for se, cw := range subs {
=======
				s.close(&wg)
				return
			}
			for se, cw := range s.subs {
>>>>>>> 12a5469... start on swarm services; move to glade
				if p.Database == se.db && p.RetentionPolicy == se.rp {
					select {
					case cw.writeRequests <- p:
					default:
<<<<<<< HEAD
						s.failures.Add(1)
=======
						atomic.AddInt64(&s.stats.WriteFailures, 1)
>>>>>>> 12a5469... start on swarm services; move to glade
					}
				}
			}
		}
	}
}

<<<<<<< HEAD
func (s *Service) updateSubs(subs map[subEntry]chanWriter, wg *sync.WaitGroup) error {
=======
// close closes the existing channel writers
func (s *Service) close(wg *sync.WaitGroup) {
	s.subMu.Lock()
	defer s.subMu.Unlock()

	for _, cw := range s.subs {
		cw.Close()
	}
	// Wait for them to finish
	wg.Wait()
	s.subs = nil
}

func (s *Service) updateSubs(wg *sync.WaitGroup) error {
	s.subMu.Lock()
	defer s.subMu.Unlock()

	if s.subs == nil {
		s.subs = make(map[subEntry]chanWriter)
	}

>>>>>>> 12a5469... start on swarm services; move to glade
	dbis := s.MetaClient.Databases()
	allEntries := make(map[subEntry]bool, 0)
	// Add in new subscriptions
	for _, dbi := range dbis {
		for _, rpi := range dbi.RetentionPolicies {
			for _, si := range rpi.Subscriptions {
				se := subEntry{
					db:   dbi.Name,
					rp:   rpi.Name,
					name: si.Name,
				}
				allEntries[se] = true
<<<<<<< HEAD
				if _, ok := subs[se]; ok {
=======
				if _, ok := s.subs[se]; ok {
>>>>>>> 12a5469... start on swarm services; move to glade
					continue
				}
				sub, err := s.createSubscription(se, si.Mode, si.Destinations)
				if err != nil {
					return err
				}
				cw := chanWriter{
					writeRequests: make(chan *coordinator.WritePointsRequest, 100),
					pw:            sub,
<<<<<<< HEAD
					failures:      s.failures,
					pointsWritten: s.pointsWritten,
=======
					pointsWritten: &s.stats.PointsWritten,
					failures:      &s.stats.WriteFailures,
>>>>>>> 12a5469... start on swarm services; move to glade
					logger:        s.Logger,
				}
				wg.Add(1)
				go func() {
					defer wg.Done()
					cw.Run()
				}()
<<<<<<< HEAD
				subs[se] = cw
=======
				s.subs[se] = cw
>>>>>>> 12a5469... start on swarm services; move to glade
				s.Logger.Println("added new subscription for", se.db, se.rp)
			}
		}
	}

	// Remove deleted subs
<<<<<<< HEAD
	for se := range subs {
		if !allEntries[se] {
			// Close the chanWriter
			subs[se].Close()

			// Remove it from the set
			delete(subs, se)
=======
	for se := range s.subs {
		if !allEntries[se] {
			// Close the chanWriter
			s.subs[se].Close()

			// Remove it from the set
			delete(s.subs, se)
>>>>>>> 12a5469... start on swarm services; move to glade
			s.Logger.Println("deleted old subscription for", se.db, se.rp)
		}
	}

	return nil
}

// Creates a PointsWriter from the given URL
func (s *Service) newPointsWriter(u url.URL) (PointsWriter, error) {
	switch u.Scheme {
	case "udp":
		return NewUDP(u.Host), nil
	case "http", "https":
		return NewHTTP(u.String(), time.Duration(s.conf.HTTPTimeout))
	default:
		return nil, fmt.Errorf("unknown destination scheme %s", u.Scheme)
	}
}

// Sends WritePointsRequest to a PointsWriter received over a channel.
type chanWriter struct {
	writeRequests chan *coordinator.WritePointsRequest
	pw            PointsWriter
<<<<<<< HEAD
	pointsWritten *expvar.Int
	failures      *expvar.Int
=======
	pointsWritten *int64
	failures      *int64
>>>>>>> 12a5469... start on swarm services; move to glade
	logger        *log.Logger
}

// Close the chanWriter
func (c chanWriter) Close() {
	close(c.writeRequests)
}

func (c chanWriter) Run() {
	for wr := range c.writeRequests {
		err := c.pw.WritePoints(wr)
		if err != nil {
			c.logger.Println(err)
<<<<<<< HEAD
			c.failures.Add(1)
		} else {
			c.pointsWritten.Add(int64(len(wr.Points)))
=======
			atomic.AddInt64(c.failures, 1)
		} else {
			atomic.AddInt64(c.pointsWritten, int64(len(wr.Points)))
>>>>>>> 12a5469... start on swarm services; move to glade
		}
	}
}

<<<<<<< HEAD
=======
// Statistics returns statistics for periodic monitoring.
func (c chanWriter) Statistics(tags map[string]string) []models.Statistic {
	if m, ok := c.pw.(monitor.Reporter); ok {
		return m.Statistics(tags)
	}
	return []models.Statistic{}
}

>>>>>>> 12a5469... start on swarm services; move to glade
// BalanceMode sets what balance mode to use on a subscription.
// valid options are currently ALL or ANY
type BalanceMode int

//ALL is a Balance mode option
const (
	ALL BalanceMode = iota
	ANY
)

<<<<<<< HEAD
// balances writes across PointsWriters according to BalanceMode
type balancewriter struct {
	bm       BalanceMode
	writers  []PointsWriter
	statMaps []*expvar.Map
	i        int
=======
type writerStats struct {
	dest          string
	failures      int64
	pointsWritten int64
}

// balances writes across PointsWriters according to BalanceMode
type balancewriter struct {
	bm      BalanceMode
	writers []PointsWriter
	stats   []writerStats
	tags    map[string]string
	i       int
>>>>>>> 12a5469... start on swarm services; move to glade
}

func (b *balancewriter) WritePoints(p *coordinator.WritePointsRequest) error {
	var lastErr error
	for range b.writers {
		// round robin through destinations.
		i := b.i
		w := b.writers[i]
		b.i = (b.i + 1) % len(b.writers)

		// write points to destination.
		err := w.WritePoints(p)
		if err != nil {
			lastErr = err
<<<<<<< HEAD
			b.statMaps[i].Add(statWriteFailures, 1)
		} else {
			b.statMaps[i].Add(statPointsWritten, int64(len(p.Points)))
=======
			atomic.AddInt64(&b.stats[i].failures, 1)
		} else {
			atomic.AddInt64(&b.stats[i].pointsWritten, int64(len(p.Points)))
>>>>>>> 12a5469... start on swarm services; move to glade
			if b.bm == ANY {
				break
			}
		}
	}
	return lastErr
}
<<<<<<< HEAD
=======

// Statistics returns statistics for periodic monitoring.
func (b *balancewriter) Statistics(tags map[string]string) []models.Statistic {
	tags = models.Tags(tags).Merge(b.tags)

	statistics := make([]models.Statistic, len(b.stats))
	for i := range b.stats {
		statistics[i] = models.Statistic{
			Name: "subscriber",
			Tags: models.Tags(tags).Merge(map[string]string{"destination": b.stats[i].dest}),
			Values: map[string]interface{}{
				statPointsWritten: atomic.LoadInt64(&b.stats[i].pointsWritten),
				statWriteFailures: atomic.LoadInt64(&b.stats[i].failures),
			},
		}
	}
	return statistics
}
>>>>>>> 12a5469... start on swarm services; move to glade
