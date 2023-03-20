package genericsync_test

import (
	"bytes"
	"reflect"
	"sync"
	"testing"
	"time"

	"github.com/ngicks/genericsync"
	"github.com/stretchr/testify/require"
)

type kv[K, V comparable] struct {
	Key   K
	Value V
}

func mapTestSet[K, V comparable](t *testing.T, m *genericsync.Map[K, V], keyValues ...kv[K, V]) {
	require := require.New(t)

	if len(keyValues) < 4 {
		panic("keyValues must be length of 3 or more")
	}

	lastEle := keyValues[len(keyValues)-1]
	keyValues = keyValues[:len(keyValues)-1]

	// Store
	for _, kv := range keyValues {
		m.Store(kv.Key, kv.Value)
	}

	// Load
	for _, kv := range keyValues {
		v, ok := m.Load(kv.Key)
		require.True(ok, "Load")
		if !reflect.DeepEqual(v, kv.Value) {
			t.Fatalf("mismatched stored value, stored: %v, loaded: %v", kv.Value, v)
		}
	}
	v, ok := m.Load(lastEle.Key)
	require.False(ok, "Load")
	if !reflect.ValueOf(v).IsZero() {
		t.Fatalf("mismatched value, must be zero value, but loaded: %v", v)
	}

	// Range
	m.Range(func(key K, value V) bool {
		for _, kv := range keyValues {
			if reflect.DeepEqual(key, kv.Key) && reflect.DeepEqual(value, kv.Value) {
				return true
			}
		}
		t.Fatalf("Range passing incorrect value")
		return true
	})

	// LoadOrStore
	if v, loaded := m.LoadOrStore(keyValues[0].Key, keyValues[0].Value); !reflect.DeepEqual(v, keyValues[0].Value) || !loaded {
		t.Fatalf("must be stored but could not load")
	}
	if v, loaded := m.LoadOrStore(keyValues[0].Key, keyValues[1].Value); !reflect.DeepEqual(v, keyValues[0].Value) || !loaded {
		t.Fatalf("must be stored but could not load")
	}
	if v, _ := m.LoadOrStore(lastEle.Key, lastEle.Value); !reflect.DeepEqual(v, lastEle.Value) {
		t.Fatalf("mismatched stored value, stored: %v, loaded: %v", lastEle.Value, v)
	}

	// LoadAndDelete
	if v, loaded := m.LoadAndDelete(keyValues[0].Key); !reflect.DeepEqual(v, keyValues[0].Value) || !loaded {
		t.Fatalf("mismatched stored value, stored: %v, loaded: %v", keyValues[0].Value, v)
	}
	if v, loaded := m.LoadAndDelete(keyValues[0].Key); !reflect.ValueOf(v).IsZero() || loaded {
		t.Fatalf("must be deleted already and loaded value must be zero. loaded: %v", v)
	}
	if _, loaded := m.Load(keyValues[0].Key); loaded {
		t.Fatalf("must be deleted")
	}

	// Delete
	if _, loaded := m.Load(keyValues[1].Key); !loaded {
		t.Fatalf("must be stored")
	}
	m.Delete(keyValues[1].Key)
	if _, loaded := m.Load(keyValues[1].Key); loaded {
		t.Fatalf("must be deleted")
	}

	//CompareAndDelete
	if _, loaded := m.Load(keyValues[2].Key); !loaded {
		t.Fatalf("must be stored")
	}
	require.False(m.CompareAndDelete(keyValues[2].Key, lastEle.Value))
	if _, loaded := m.Load(keyValues[2].Key); !loaded {
		t.Fatalf("must be stored")
	}
	require.True(m.CompareAndDelete(keyValues[2].Key, keyValues[2].Value))
	if _, loaded := m.Load(keyValues[2].Key); loaded {
		t.Fatalf("must be deleted")
	}

	//CompareAndSwap
	if v, loaded := m.Load(keyValues[3].Key); !loaded || !reflect.DeepEqual(v, keyValues[3].Value) {
		t.Fatalf("must be stored, unchanged")
	}

	require.False(
		m.CompareAndSwap(keyValues[3].Key, keyValues[2].Value, lastEle.Value),
	)

	if v, loaded := m.Load(keyValues[3].Key); !loaded || !reflect.DeepEqual(v, keyValues[3].Value) {
		t.Fatalf("must be stored, unchanged")
	}
	require.True(
		m.CompareAndSwap(keyValues[3].Key, keyValues[3].Value, lastEle.Value),
	)
	if v, loaded := m.Load(keyValues[3].Key); !loaded || reflect.DeepEqual(v, keyValues[3].Value) {
		t.Fatalf("must be stored but changed")
	}
}

func TestMap(t *testing.T) {
	m := genericsync.Map[string, time.Time]{}
	now := time.Now()
	mapTestSet(t, &m,
		[]kv[string, time.Time]{
			{"foo", now},
			{"bar", now.Add(time.Hour)},
			{"baz", now.Add(2 * time.Hour)},
			{"qux", now.Add(3 * time.Hour)},
			{"quux", now.Add(4 * time.Hour)},
		}...,
	)
}

func TestMapRace(t *testing.T) {
	m := genericsync.Map[time.Time, *bytes.Buffer]{}
	now := time.Now()

	callAllMethodsInRandomOrder := func() {
		keyOrder := map[int]any{
			0: nil,
			1: nil,
			2: nil,
			3: nil,
			4: nil,
			5: nil,
			6: nil,
			7: nil,
		}
		for key := range keyOrder {
			switch key {
			case 0:
				m.Delete(now)
			case 1:
				m.Load(now)
			case 2:
				m.LoadAndDelete(now)
			case 3:
				m.LoadOrStore(now, new(bytes.Buffer))
			case 4:
				m.Range(func(key time.Time, value *bytes.Buffer) bool {
					return true
				})
			case 5:
				m.Store(now, new(bytes.Buffer))
			case 6:
				m.CompareAndDelete(now, new(bytes.Buffer))
			case 7:
				m.CompareAndSwap(now, new(bytes.Buffer), bytes.NewBuffer([]byte(`foo`)))
			}
		}
	}

	wg := sync.WaitGroup{}
	for i := 0; i < 1000; i++ {
		wg.Add(1)
		go func() {
			callAllMethodsInRandomOrder()
			wg.Done()
		}()
	}
	wg.Wait()
}
