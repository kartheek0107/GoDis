package store

import (
	"hash/fnv"
	"math"
	"math/bits"
)

// HyperLogLog Implementation for Cardinality Estimation

type HyperLogLog struct {
	buckets [64]uint8
}

func NewHyperLogLog() *HyperLogLog {
	return &HyperLogLog{}
}

func (hll *HyperLogLog) Add(val string) {
	h := fnv.New64a()
	h.Write([]byte(val))
	hash := h.Sum64()
	idx := hash >> 58 // top 6 bits for 64 buckets
	zeros := uint8(bits.LeadingZeros64(hash<<6)) + 1
	if zeros > hll.buckets[idx] {
		hll.buckets[idx] = zeros
	}
}

func (hll *HyperLogLog) Count() int {
	var sum float64
	zeros := 0
	for _, v := range hll.buckets {
		sum += 1.0 / float64(uint64(1)<<v)
		if v == 0 {
			zeros++
		}
	}
	alpha := 0.709
	m := 64.0
	estimate := alpha * m * m / sum
	if estimate <= 2.5*m {
		if zeros > 0 {
			estimate = m * math.Log(m/float64(zeros))
		}
	}
	return int(estimate)
}

func (s *Store) PfAdd(key string, elements ...string) (int, error) {
	s.mux.Lock()
	defer s.mux.Unlock()

	val, ok := s.Data[key]
	var hll *HyperLogLog
	if !ok {
		hll = NewHyperLogLog()
		s.Data[key] = hll
	} else {
		existingHLL, isHLL := val.(*HyperLogLog)
		if !isHLL {
			return -1, nil // WRONGTYPE
		}
		hll = existingHLL
	}

	// Simply add, ideally we'd track if internal state changed to return 1 or 0
	// For simplicity, we just add and return 1 if we added at least 1 element
	for _, elem := range elements {
		hll.Add(elem)
	}

	if len(elements) > 0 {
		return 1, nil
	}
	return 0, nil
}

func (s *Store) PfCount(key string) (int, error) {
	s.mux.RLock()
	defer s.mux.RUnlock()

	val, ok := s.Data[key]
	if !ok {
		return 0, nil
	}

	hll, isHLL := val.(*HyperLogLog)
	if !isHLL {
		return -1, nil
	}

	return hll.Count(), nil
}

// Bloom Filter Implementation for Membership Testing

type BloomFilter struct {
	bitset []uint64
	size   uint64
	hashes uint
}

func NewBloomFilter(size uint64, hashes uint) *BloomFilter {
	words := (size + 63) / 64
	return &BloomFilter{
		bitset: make([]uint64, words),
		size:   size,
		hashes: hashes,
	}
}

func (bf *BloomFilter) getHashes(val string) []uint64 {
	h := fnv.New64a()
	h.Write([]byte(val))
	hash1 := h.Sum64()

	h2 := fnv.New32a()
	h2.Write([]byte(val))
	hash2 := uint64(h2.Sum32())

	res := make([]uint64, bf.hashes)
	for i := uint(0); i < bf.hashes; i++ {
		res[i] = (hash1 + uint64(i)*hash2) % bf.size
	}
	return res
}

func (bf *BloomFilter) Add(val string) {
	hashes := bf.getHashes(val)
	for _, h := range hashes {
		wordIdx := h / 64
		bitIdx := h % 64
		bf.bitset[wordIdx] |= (1 << bitIdx)
	}
}

func (bf *BloomFilter) Exists(val string) bool {
	hashes := bf.getHashes(val)
	for _, h := range hashes {
		wordIdx := h / 64
		bitIdx := h % 64
		if (bf.bitset[wordIdx] & (1 << bitIdx)) == 0 {
			return false
		}
	}
	return true
}

func (s *Store) BfAdd(key string, val string) (int, error) {
	s.mux.Lock()
	defer s.mux.Unlock()

	v, ok := s.Data[key]
	var bf *BloomFilter
	if !ok {
		// Default size 10000, 5 hashes
		bf = NewBloomFilter(10000, 5)
		s.Data[key] = bf
	} else {
		existingBF, isBF := v.(*BloomFilter)
		if !isBF {
			return -1, nil // WRONGTYPE
		}
		bf = existingBF
	}

	if bf.Exists(val) {
		return 0, nil
	}
	bf.Add(val)
	return 1, nil
}

func (s *Store) BfExists(key string, val string) (int, error) {
	s.mux.RLock()
	defer s.mux.RUnlock()

	v, ok := s.Data[key]
	if !ok {
		return 0, nil
	}

	bf, isBF := v.(*BloomFilter)
	if !isBF {
		return -1, nil
	}

	if bf.Exists(val) {
		return 1, nil
	}
	return 0, nil
}
