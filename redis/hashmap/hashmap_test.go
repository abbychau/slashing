package HashMap

import (
	"encoding/json"
	"strconv"
	"sync"
	"testing"
	"time"
)

func TestNewHashMap(t *testing.T) {
	hm := New()
	batch := 1000000
	start := time.Now().UnixNano()
	for i := 0; i < batch; i++ {
		hm.Set(i, i)
	}
	end := time.Now().UnixNano()
	t.Log("set:", (end-start)/1e6)

	start = time.Now().UnixNano()
	for i := 0; i < batch; i++ {
		v, e := hm.Get(i)
		if !e || v != i {
			t.Fatal("data err ", i)
		}
	}
	end = time.Now().UnixNano()
	t.Log("get:", (end-start)/1e6)
}

func TestHashMapCorrectness(t *testing.T) {
	hm := New()
	batch := 1000000
	start := time.Now().UnixNano()
	for i := 0; i < batch; i++ {
		hm.Set(i, i)
	}
	for i := 0; i < batch; i++ {
		v, _ := hm.Get(i)
		v2 := hm.Set(i, v.(int)+1)
		if v2 != i {
			t.Fatalf("v2 val : %v", v2)
		}
	}
	for i := 0; i < batch; i++ {
		v, e := hm.Get(i)
		if !e || v != i+1 {
			t.Fatal("data err ", i)
		}
	}
	end := time.Now().UnixNano()
	t.Log("get:", (end-start)/1e6)
}

func TestNewHashMap_Sync(t *testing.T) {
	hm := New()
	batch := 100000
	wg := sync.WaitGroup{}
	wg.Add(batch)
	for i := 0; i < batch; i++ {
		n := i
		go func() {
			hm.Set(strconv.Itoa(n), n)
			wg.Done()
		}()
	}
	wg.Wait()
	t.Log(hm.size)
	if hm.size != int64(batch) {
		t.Fatal("TestNewHashMap_Sync SET ERR")
	}

	wg.Add(batch / 2)
	for i := 0; i < batch/2; i++ {
		n := i
		go func() {
			hm.Del(strconv.Itoa(n))
			wg.Done()
		}()
	}
	wg.Wait()
	if hm.size != int64(batch-batch/2) {
		t.Fatal("TestNewHashMap_Sync DEL ERR")
	}
}

func TestHashMap_MarshalJSON(t *testing.T) {
	m := New()
	m.Set("abc", "haha")
	m.Set(1, 2)
	m.Set("m", map[string]string{
		"hello": "world",
	})
	b, err := json.Marshal(m)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("%s\n", b)
}

func TestHashMap_UnmarshalJSON(t *testing.T) {
	jsonStr := "{\"1\":2,\"abc\":\"haha\",\"m\":{\"hello\":\"world\"}}"
	m := New()
	err := json.Unmarshal([]byte(jsonStr), m)
	if err != nil {
		t.Fatal(err)
	}
	b, err := json.Marshal(m)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(string(b))
	t.Log(m.Get("1"))
}

func TestHashMap_BinaryMarshaling(t *testing.T) {
	jsonStr := "{\"1\":2,\"abc\":\"haha\",\"m\":{\"hello\":\"world\"}}"
	m := New()
	m.FromJSON([]byte(jsonStr))
	t.Log(m.Get("1"))
	myBin, err := m.ToBinary()
	if err != nil {
		t.Fatalf("%v", err)
	}
	m2 := NewFromBinary(myBin)
	v, _ := m2.Get("1")
	if v != 2.0 {
		t.Errorf("not 2.0, but %v", v)
	}

}
