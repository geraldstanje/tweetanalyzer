package classifier

import (
  "sort"
)

type sortedmap struct {
}

func NewSortedMap() *sortedmap {
  return &sortedmap{}
}

type Dict map[string]int

// A slice of Pairs that implements sort.Interface to sort by Value.
type PairList []Pair

// A data structure to hold a key/value pair.
type Pair struct {
  Key   string
  Value int
}

func (p PairList) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }
func (p PairList) Len() int           { return len(p) }
func (p PairList) Less(i, j int) bool { return p[i].Value < p[j].Value }

// A function returns a by value sorted map
func (s *sortedmap) SortByValue(m *Dict) {
  p := make(PairList, len(*m))
  i := 0
  for k, v := range *m {
    p[i] = Pair{k, v}
    i++
  }
  sort.Sort(p)

  n := make(Dict)
  for _, v := range p {
    n[v.Key] = v.Value
  }
  *m = n
}