package store

// Heatmap represents a 2D grid storing integer weights (e.g. tracking clicks/activity).
type Heatmap struct {
	grid map[int]map[int]int
}

func NewHeatmap() *Heatmap {
	return &Heatmap{
		grid: make(map[int]map[int]int),
	}
}

func (hm *Heatmap) Set(x, y, val int) {
	if _, ok := hm.grid[x]; !ok {
		hm.grid[x] = make(map[int]int)
	}
	hm.grid[x][y] = val
}

func (hm *Heatmap) Get(x, y int) int {
	if row, ok := hm.grid[x]; ok {
		if val, ok2 := row[y]; ok2 {
			return val
		}
	}
	return 0
}

func (s *Store) HeatmapSet(key string, x, y, val int) error {
	s.mux.Lock()
	defer s.mux.Unlock()

	v, ok := s.Data[key]
	var hm *Heatmap
	if !ok {
		hm = NewHeatmap()
		s.Data[key] = hm
	} else {
		existingHM, isHM := v.(*Heatmap)
		if !isHM {
			return nil // Should return error, will return nil and handle in server
		}
		hm = existingHM
	}

	hm.Set(x, y, val)
	return nil
}

func (s *Store) HeatmapGet(key string, x, y int) (int, bool) {
	s.mux.RLock()
	defer s.mux.RUnlock()

	v, ok := s.Data[key]
	if !ok {
		return 0, false
	}

	hm, isHM := v.(*Heatmap)
	if !isHM {
		return 0, false // WRONGTYPE
	}

	return hm.Get(x, y), true
}
