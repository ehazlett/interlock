// Copyright 2015 The etcd Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package compactor

import (
	"reflect"
	"testing"
	"time"

	pb "github.com/coreos/etcd/etcdserver/etcdserverpb"
	"github.com/coreos/etcd/pkg/testutil"
	"github.com/jonboulle/clockwork"
	"golang.org/x/net/context"
)

func TestPeriodic(t *testing.T) {
	fc := clockwork.NewFakeClock()
<<<<<<< HEAD
=======
	rg := &fakeRevGetter{testutil.NewRecorderStream(), 0}
>>>>>>> 12a5469... start on swarm services; move to glade
	compactable := &fakeCompactable{testutil.NewRecorderStream()}
	tb := &Periodic{
		clock:        fc,
		periodInHour: 1,
<<<<<<< HEAD
		rg:           &fakeRevGetter{},
=======
		rg:           rg,
>>>>>>> 12a5469... start on swarm services; move to glade
		c:            compactable,
	}

	tb.Run()
	defer tb.Stop()

	n := int(time.Hour / checkCompactionInterval)
	for i := 0; i < 3; i++ {
		for j := 0; j < n; j++ {
<<<<<<< HEAD
			time.Sleep(5 * time.Millisecond)
=======
			rg.Wait(1)
>>>>>>> 12a5469... start on swarm services; move to glade
			fc.Advance(checkCompactionInterval)
		}

		a, err := compactable.Wait(1)
		if err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(a[0].Params[0], &pb.CompactionRequest{Revision: int64(i*n) + 1}) {
			t.Errorf("compact request = %v, want %v", a[0].Params[0], &pb.CompactionRequest{Revision: int64(i*n) + 1})
		}
	}
}

func TestPeriodicPause(t *testing.T) {
	fc := clockwork.NewFakeClock()
	compactable := &fakeCompactable{testutil.NewRecorderStream()}
<<<<<<< HEAD
	tb := &Periodic{
		clock:        fc,
		periodInHour: 1,
		rg:           &fakeRevGetter{},
=======
	rg := &fakeRevGetter{testutil.NewRecorderStream(), 0}
	tb := &Periodic{
		clock:        fc,
		periodInHour: 1,
		rg:           rg,
>>>>>>> 12a5469... start on swarm services; move to glade
		c:            compactable,
	}

	tb.Run()
	tb.Pause()

	n := int(time.Hour / checkCompactionInterval)
	for i := 0; i < 3*n; i++ {
<<<<<<< HEAD
		time.Sleep(5 * time.Millisecond)
=======
		rg.Wait(1)
>>>>>>> 12a5469... start on swarm services; move to glade
		fc.Advance(checkCompactionInterval)
	}

	select {
	case a := <-compactable.Chan():
		t.Fatalf("unexpected action %v", a)
	case <-time.After(10 * time.Millisecond):
	}

	tb.Resume()
<<<<<<< HEAD
=======
	rg.Wait(1)
>>>>>>> 12a5469... start on swarm services; move to glade
	fc.Advance(checkCompactionInterval)

	a, err := compactable.Wait(1)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(a[0].Params[0], &pb.CompactionRequest{Revision: int64(2*n) + 2}) {
		t.Errorf("compact request = %v, want %v", a[0].Params[0], &pb.CompactionRequest{Revision: int64(2*n) + 2})
	}
}

type fakeCompactable struct {
	testutil.Recorder
}

func (fc *fakeCompactable) Compact(ctx context.Context, r *pb.CompactionRequest) (*pb.CompactionResponse, error) {
	fc.Record(testutil.Action{Name: "c", Params: []interface{}{r}})
	return &pb.CompactionResponse{}, nil
}

type fakeRevGetter struct {
<<<<<<< HEAD
=======
	testutil.Recorder
>>>>>>> 12a5469... start on swarm services; move to glade
	rev int64
}

func (fr *fakeRevGetter) Rev() int64 {
<<<<<<< HEAD
=======
	fr.Record(testutil.Action{Name: "g"})
>>>>>>> 12a5469... start on swarm services; move to glade
	fr.rev++
	return fr.rev
}
