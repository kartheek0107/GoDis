package store

// Lists (Deque implementation)
func (s *Store) LPush(key string, values ...string) int {
	s.mux.Lock()
	defer s.mux.Unlock()

	val, ok := s.Data[key]
	var list []string
	if !ok {
		list = make([]string, 0)
	} else {
		existingList, isList := val.([]string)
		if !isList {
			return -1 // Indicates WRONGTYPE
		}
		list = existingList
	}

	// Prepend
	newList := make([]string, len(values))
	// In LPUSH, multiple values are pushed from left to right.
	// LPUSH mylist a b c results in c b a
	for i, v := range values {
		newList[len(values)-1-i] = v
	}
	list = append(newList, list...)
	s.Data[key] = list
	return len(list)
}

func (s *Store) RPush(key string, values ...string) int {
	s.mux.Lock()
	defer s.mux.Unlock()

	val, ok := s.Data[key]
	var list []string
	if !ok {
		list = make([]string, 0)
	} else {
		existingList, isList := val.([]string)
		if !isList {
			return -1
		}
		list = existingList
	}

	list = append(list, values...)
	s.Data[key] = list
	return len(list)
}

func (s *Store) LPop(key string) (string, bool) {
	s.mux.Lock()
	defer s.mux.Unlock()

	val, ok := s.Data[key]
	if !ok {
		return "", false
	}

	list, isList := val.([]string)
	if !isList || len(list) == 0 {
		return "", false
	}

	popped := list[0]
	if len(list) == 1 {
		delete(s.Data, key)
	} else {
		s.Data[key] = list[1:]
	}
	return popped, true
}

func (s *Store) RPop(key string) (string, bool) {
	s.mux.Lock()
	defer s.mux.Unlock()

	val, ok := s.Data[key]
	if !ok {
		return "", false
	}

	list, isList := val.([]string)
	if !isList || len(list) == 0 {
		return "", false
	}

	popped := list[len(list)-1]
	if len(list) == 1 {
		delete(s.Data, key)
	} else {
		s.Data[key] = list[:len(list)-1]
	}
	return popped, true
}

func (s *Store) LRange(key string, start, stop int) ([]string, bool) {
	s.mux.RLock()
	defer s.mux.RUnlock()

	val, ok := s.Data[key]
	if !ok {
		return nil, false
	}

	list, isList := val.([]string)
	if !isList {
		return nil, false
	}

	length := len(list)
	if start < 0 {
		start = length + start
	}
	if stop < 0 {
		stop = length + stop
	}
	if start < 0 {
		start = 0
	}
	if start >= length {
		return []string{}, true
	}
	if stop >= length {
		stop = length - 1
	}
	if start > stop {
		return []string{}, true
	}

	return list[start:stop+1], true
}

// Hashes
func (s *Store) HSet(key, field, value string) int {
	s.mux.Lock()
	defer s.mux.Unlock()

	val, ok := s.Data[key]
	var hash map[string]string
	if !ok {
		hash = make(map[string]string)
	} else {
		existingHash, isHash := val.(map[string]string)
		if !isHash {
			return -1
		}
		hash = existingHash
	}

	_, fieldExists := hash[field]
	hash[field] = value
	s.Data[key] = hash

	if fieldExists {
		return 0
	}
	return 1
}

func (s *Store) HGet(key, field string) (string, bool) {
	s.mux.RLock()
	defer s.mux.RUnlock()

	val, ok := s.Data[key]
	if !ok {
		return "", false
	}

	hash, isHash := val.(map[string]string)
	if !isHash {
		return "", false
	}

	fieldVal, fieldOk := hash[field]
	return fieldVal, fieldOk
}

func (s *Store) HGetAll(key string) (map[string]string, bool) {
	s.mux.RLock()
	defer s.mux.RUnlock()

	val, ok := s.Data[key]
	if !ok {
		return nil, false
	}

	hash, isHash := val.(map[string]string)
	if !isHash {
		return nil, false
	}

	// Return a copy to avoid concurrent map read/write if used elsewhere without lock
	res := make(map[string]string)
	for k, v := range hash {
		res[k] = v
	}
	return res, true
}

// Sets
func (s *Store) SAdd(key string, members ...string) int {
	s.mux.Lock()
	defer s.mux.Unlock()

	val, ok := s.Data[key]
	var set map[string]struct{}
	if !ok {
		set = make(map[string]struct{})
	} else {
		existingSet, isSet := val.(map[string]struct{})
		if !isSet {
			return -1
		}
		set = existingSet
	}

	added := 0
	for _, member := range members {
		if _, exists := set[member]; !exists {
			set[member] = struct{}{}
			added++
		}
	}
	s.Data[key] = set
	return added
}

func (s *Store) SIsMember(key, member string) (bool, bool) {
	s.mux.RLock()
	defer s.mux.RUnlock()

	val, ok := s.Data[key]
	if !ok {
		return false, false
	}

	set, isSet := val.(map[string]struct{})
	if !isSet {
		return false, false
	}

	_, exists := set[member]
	return exists, true
}

func (s *Store) SMembers(key string) ([]string, bool) {
	s.mux.RLock()
	defer s.mux.RUnlock()

	val, ok := s.Data[key]
	if !ok {
		return nil, false
	}

	set, isSet := val.(map[string]struct{})
	if !isSet {
		return nil, false
	}

	members := make([]string, 0, len(set))
	for k := range set {
		members = append(members, k)
	}
	return members, true
}
