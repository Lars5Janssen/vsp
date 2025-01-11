package utils

type IntSet map[int]struct{}

func NewIntSet() IntSet {
	return make(IntSet)
}

func (s IntSet) Add(i int) {
	s[i] = struct{}{}
}

func (s IntSet) Remove(i int) {
	delete(s, i)
}

func (s IntSet) Contains(i int) bool {
	_, ok := s[i]
	return ok
}

func (s IntSet) Size() int {
	return len(s)
}
