package classifier

import (
  "sort"
  "strings"
)

type Dict map[string]int

// A slice of Pairs that implements sort.Interface to sort by Value.
type PairList []Pair

// A data structure to hold a key/value pair.
type Pair struct {
  Key   string
  Value int
}

func (s Dict) Add(key string, value int) {
  s[key] = value
}

func (s Dict) Peek(key string) (int, bool) {
  ret, ok := s[key]
  return ret, ok
}

func (p PairList) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }
func (p PairList) Len() int           { return len(p) }
func (p PairList) Less(i, j int) bool { return p[i].Value < p[j].Value }

// A function returns a by value sorted map
func SortMapByValue(m *Dict) {
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

func CreateBigrams(s string) []string {
  i := 0
  j := strings.Index(s, " ")
  if j < 0 {
    return nil
  }
  j += 1
  var result []string
  for {
    k := strings.Index(s[j:], " ")
    if k < 0 {
      result = append(result, s[i:])
      break
    }
    result = append(result, s[i:j+k])
    i = j
    j += k + 1
  }
  return result
}

func strcmp(a, b string) int {
  min := len(b)
  if len(a) < len(b) {
    min = len(a)
  }
  diff := 0
  for i := 0; i < min && diff == 0; i++ {
    diff = int(a[i]) - int(b[i])
  }
  if diff == 0 {
    diff = len(a) - len(b)
  }
  return diff
}