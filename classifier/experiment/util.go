package classifier

import (
	"sort"
	"strconv"
	"strings"
)

type Dict map[string]int
type Dict2 map[string]float64

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

func (s Dict2) Add(key string, value float64) {
	s[key] = value
}

func (s Dict2) Peek(key string) (float64, bool) {
	ret, ok := s[key]
	return ret, ok
}

func stringToInt(str string) int {
	intVal, _ := strconv.ParseInt(str, 0, 64)
	return int(intVal)
}

func stringToFloat(str string) float64 {
	floatVal, _ := strconv.ParseFloat(str, 64)
	return floatVal
}

func intToString(num int) string {
	str := strconv.FormatInt(int64(num), 10)
	return str
}

func floatToString(num float64) string {
	str := strconv.FormatFloat(num, 'f', 0, 64)
	return str
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

func buildDb(ngram_db map[string]int, txt string) {
	txt = strings.TrimSpace(txt)
	if _, ok := ngram_db[txt]; ok {
		ngram_db[txt]++
	} else {
		ngram_db[txt] = 1
	}
}

func CreateNgrams(txt string, n int) map[string]int {
	ngram_db := make(Dict)

	words := strings.Split(txt, " ")

	i := 0
	limit := len(words) - (n - 1)
	for i < limit {
		buildDb(ngram_db, strings.Join(words[i:(i+n)], " "))
		i++
	}
	return ngram_db
}
