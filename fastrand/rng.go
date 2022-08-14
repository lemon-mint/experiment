package fastrand

import (
	"crypto/rand"
	"encoding/binary"
	"sync"

	"v8.run/go/exp/fastrand/alg/splitmix64"
	"v8.run/go/exp/util/noescape"
)

type RNG struct {
	state [2]uint64
}

func refill(r *RNG) {
	var data [16]byte
	_, err := noescape.Read(rand.Reader, data[:])
	if err != nil {
		// If crypto/rand fails, we'll just use the runtime.fastrand implementation.
		binary.LittleEndian.PutUint32(data[0:4], runtime_fastrand())
		binary.LittleEndian.PutUint32(data[4:8], runtime_fastrand())
		binary.LittleEndian.PutUint32(data[8:12], runtime_fastrand())
		binary.LittleEndian.PutUint32(data[12:16], runtime_fastrand())
	}

	r.state[0] = binary.LittleEndian.Uint64(data[0:8])
	r.state[1] = binary.LittleEndian.Uint64(data[8:16])

	// Use Splitmix64 to initialize the state.
	r.state[0] += 1757750930446974760
	r.state[1] += 7151402297004559274
	r.state[0] = splitmix64.Splitmix64(&r.state[0])
	r.state[1] = splitmix64.Splitmix64(&r.state[1])
}

func newRNG() *RNG {
	r := new(RNG)
	refill(r)
	return r
}

var rngPool sync.Pool = sync.Pool{
	New: func() interface{} {
		return newRNG()
	},
}

func AcquireRNG() *RNG {
	return rngPool.Get().(*RNG)
}

func ReleaseRNG(r *RNG) {
	rngPool.Put(r)
}

func WithSeed(seed uint64) *RNG {
	r := AcquireRNG()
	r.SetSeed(seed)
	return r
}

func (r *RNG) SetSeed(seed uint64) {
	r.state[0] = splitmix64.Splitmix64(&seed)
	r.state[1] = splitmix64.Splitmix64(&seed)
}

// Release Put the RNG back into the pool.
// After calling this, the RNG is invalid and should not be used.
func (r *RNG) Release() {
	ReleaseRNG(r)
}

// Refill Initialize the RNG with a new seed.
func (r *RNG) Refill() {
	refill(r)
}
