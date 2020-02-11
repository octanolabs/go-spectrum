package syncronizer_test

import (
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/octanolabs/go-spectrum/models"

	log "github.com/sirupsen/logrus"

	"github.com/octanolabs/go-spectrum/rpc"

	json "github.com/json-iterator/go"

	"github.com/octanolabs/go-spectrum/crawler/syncronizer"
)

var rpcClient *rpc.RPCClient

func TestMain(m *testing.M) {
	var c *rpc.Config

	rpcCfg := []byte(`{
		"type": "http",
    	"endpoint": "https://rpc.octano.dev"
	}`)

	err := json.Unmarshal(rpcCfg, &c)

	if err != nil {
		log.Errorf("Error unmarshaling: %v", err)
	}

	log.Printf("%+v", c)

	rpcClient = rpc.NewRPCClient(c)

	os.Exit(m.Run())
}

func fetchBlock(h uint64) models.Block {
	block, err := rpcClient.GetBlockByHeight(h)
	if err != nil {
		log.Errorf("Error getting block: %v", err)
	}
	return block
}

var testTable = []struct{ maxRoutines, routines int }{
	{5, 100},
	{10, 100},
	{25, 100},
	{50, 100},
	{100, 100},
	{100, 1000},
}

func SyncFunc(maxRoutines, routines int) bool {

	sync := syncronizer.NewSync(maxRoutines)

	for i := 0; i < routines; i++ {
		sync.AddLink(func(r *syncronizer.Routine) {
			time.Sleep(1 * time.Millisecond)
		})
	}

	return sync.Finish()
}

func BlockSyncFunc(maxRoutines, routines int) bool {

	sync := syncronizer.NewSync(maxRoutines)

	for i := 0; i < routines; i++ {
		_ = fetchBlock(uint64(i))

		sync.AddLink(func(r *syncronizer.Routine) {
			closed := r.Link()
			if closed {
				return
			}
		})
	}

	return sync.Finish()
}

func AsyncBlockSyncFunc(maxRoutines, routines int) bool {

	sync := syncronizer.NewSync(maxRoutines)

	for i := 0; i < routines; i++ {
		sync.AddLink(func(r *syncronizer.Routine) {

			_ = fetchBlock(uint64(i))

			closed := r.Link()
			if closed {
				return
			}
		})
	}

	return sync.Finish()
}

func TestSync(t *testing.T) {

	for k, v := range testTable {
		t.Logf("start test n.%v with %v routines, %v maxRoutines", k, v.routines, v.maxRoutines)

		start := time.Now()
		val := SyncFunc(v.maxRoutines, v.routines)
		end := time.Since(start)

		t.Logf("test n.%v with %v routines, %v maxRoutines took %v; val == %v", k, v.routines, v.maxRoutines, end, val)
	}

}

func TestSyncAbort(t *testing.T) {
	s := syncronizer.NewSync(10)

	for i := 0; i < 100; i++ {
		str := "routine_" + strconv.FormatInt(int64(i), 10)
		s.AddLink(func(r *syncronizer.Routine) {
			r.Link()
			if i == 50 {
				time.Sleep(1 * time.Millisecond)
				t.Log(str, "should close")
				s.Abort()
			} else {
				time.Sleep(1 * time.Millisecond)

				t.Log(str)
			}
		})
	}

	closed := s.Finish()

	if closed {
		t.Logf("syncBlock aborted successfully")
	} else {
		t.Errorf("Error, sync not aborted")
	}
}

func BenchmarkSync(b *testing.B) {

	for k, v := range testTable {

		n := strconv.FormatInt(int64(k), 10)

		b.Run("bench n."+n, func(b *testing.B) {
			b.Logf("start bench %v routines, %v maxRoutines", v.routines, v.maxRoutines)

			for i := 0; i < b.N; i++ {
				SyncFunc(v.maxRoutines, v.routines)
			}
		})
	}
}

func TestSyncBlocks(t *testing.T) {

	for k, v := range testTable {
		t.Logf("start test n.%v with %v routines, %v maxRoutines", k, v.routines, v.maxRoutines)

		start := time.Now()
		val := AsyncBlockSyncFunc(v.maxRoutines, v.routines)
		end := time.Since(start)

		t.Logf("test n.%v with %v routines, %v maxRoutines took %v; val == %v", k, v.routines, v.maxRoutines, end, val)
	}

	for k, v := range testTable {
		t.Logf("start test n.%v with %v routines, %v maxRoutines", k, v.routines, v.maxRoutines)

		start := time.Now()
		val := BlockSyncFunc(v.maxRoutines, v.routines)
		end := time.Since(start)

		t.Logf("test n.%v with %v routines, %v maxRoutines took %v; val == %v", k, v.routines, v.maxRoutines, end, val)
	}

}
