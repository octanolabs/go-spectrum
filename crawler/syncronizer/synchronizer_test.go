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

var testTable = []struct{ maxRoutines, routines, abortAt int }{
	{5, 100, 50},
	{10, 100, 50},
	{25, 100, 50},
	{50, 100, 50},
	{100, 100, 50},
	{100, 1000, 500},
}

func SyncFunc(maxRoutines, routines int) bool {

	sync := syncronizer.NewSync(maxRoutines)

	for i := 0; i < routines; i++ {
		sync.AddLink(func(r *syncronizer.Task) {
			r.Link()
			time.Sleep(1 * time.Millisecond)

		})
	}

	return sync.Finish()
}

func AbortBeforeSyncFunc(t *testing.T, maxRoutines, routines, abortAt int) {

	sync := syncronizer.NewSync(maxRoutines)

	for i := 0; i < routines; i++ {
		it := i
		sync.AddLink(func(r *syncronizer.Task) {

			//simulate io op to run before linking
			_ = fetchBlock(uint64(i))

			if it == abortAt {
				//fmt.Println("routine_"+strconv.FormatInt(int64(it), 10), "closing")
				r.AbortSync()
				//fmt.Println("routine_"+strconv.FormatInt(int64(it), 10), " should've closed")

			}

			closed := r.Link()

			if closed {
				return
			}

			time.Sleep(1 * time.Millisecond)

		})

	}

	f := sync.Finish()

	if f {
		t.Log("Sync Aborted successfully")
	} else {
		t.Fatalf("failed to abort sync")
	}

}

func AbortAfterSyncFunc(t *testing.T, maxRoutines, routines, abortAt int) {

	sync := syncronizer.NewSync(maxRoutines)

	for i := 0; i < routines; i++ {
		it := i
		sync.AddLink(func(r *syncronizer.Task) {
			closed := r.Link()

			if closed {
				return
			}

			//simulate op to run after linking
			//_ = fetchBlock(uint64(i))

			if it == abortAt {
				//fmt.Println("routine_"+strconv.FormatInt(int64(it), 10), "closing")
				r.AbortSync()
				//fmt.Println("routine_"+strconv.FormatInt(int64(it), 10), " should close")
				return
			}

			time.Sleep(1 * time.Millisecond)

		})
	}

	f := sync.Finish()

	if f {
		t.Log("Sync Aborted successfully")
	} else {
		t.Fatalf("failed to abort sync")
	}

}

func BlockSyncFunc(maxRoutines, routines int) bool {

	sync := syncronizer.NewSync(maxRoutines)

	for i := 0; i < routines; i++ {
		_ = fetchBlock(uint64(i))

		sync.AddLink(func(r *syncronizer.Task) {
			closed := r.Link()
			if closed {
				return
			}
			time.Sleep(1 * time.Millisecond)
		})
	}

	return sync.Finish()
}

func AsyncBlockSyncFunc(maxRoutines, routines int) bool {

	sync := syncronizer.NewSync(maxRoutines)

	for i := 0; i < routines; i++ {
		sync.AddLink(func(r *syncronizer.Task) {

			_ = fetchBlock(uint64(i))

			closed := r.Link()
			if closed {
				return
			}
			time.Sleep(time.Millisecond)
		})
	}

	return sync.Finish()
}

func TestSync(t *testing.T) {

	for k, v := range testTable {
		t.Run("test_"+strconv.FormatInt(int64(k), 10), func(t *testing.T) {
			t.Logf("start test n.%v with %v routines, %v maxRoutines", k, v.routines, v.maxRoutines)

			start := time.Now()
			val := SyncFunc(v.maxRoutines, v.routines)
			end := time.Since(start)

			t.Logf("test n.%v with %v routines, %v maxRoutines took %v; aborted %v", k, v.routines, v.maxRoutines, end, val)
		})
	}

}

func TestSyncAbort(t *testing.T) {
	for k, v := range testTable {
		t.Run("test_"+strconv.FormatInt(int64(k), 10), func(t *testing.T) {
			t.Logf("start test n.%v with %v routines, %v maxRoutines", k, v.routines, v.maxRoutines)

			start := time.Now()
			AbortBeforeSyncFunc(t, v.maxRoutines, v.routines, v.abortAt)
			end := time.Since(start)

			start1 := time.Now()
			AbortAfterSyncFunc(t, v.maxRoutines, v.routines, v.abortAt)
			end1 := time.Since(start1)

			t.Logf("test n.%v with %v routines, %v maxRoutines (before took %v, after took %v)", k, v.routines, v.maxRoutines, end, end1)
		})
	}
}

func TestSyncBlocks(t *testing.T) {

	for k, v := range testTable {
		t.Run("test_"+strconv.FormatInt(int64(k), 10), func(t *testing.T) {
			t.Logf("start test n.%v with %v routines, %v maxRoutines", k, v.routines, v.maxRoutines)

			start := time.Now()
			val := AsyncBlockSyncFunc(v.maxRoutines, v.routines)
			end := time.Since(start)

			t.Logf("test n.%v with %v routines, %v maxRoutines took %v; val == %v", k, v.routines, v.maxRoutines, end, val)
		})
	}

	//for k, v := range testTable {
	//	t.Logf("start test n.%v with %v routines, %v maxRoutines", k, v.routines, v.maxRoutines)
	//
	//	start := time.Now()
	//	val := BlockSyncFunc(v.maxRoutines, v.routines)
	//	end := time.Since(start)
	//
	//	t.Logf("test n.%v with %v routines, %v maxRoutines took %v; val == %v", k, v.routines, v.maxRoutines, end, val)
	//}

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
