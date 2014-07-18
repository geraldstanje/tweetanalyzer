package main

import (
	"bufio"
	"fmt"
	"github.com/sridif/gosvm"
	"io"
	"log"
	"os"
	"regexp"
	"strconv"
	"strings"
)

var myExp = regexp.MustCompile(`\s`)

type SliceData struct {
	i int
	s string
}

func stringToInt(str string) int {
	intVal, _ := strconv.ParseInt(str, 0, 64)
	return int(intVal)
}

func intToString(input_num int) string {
	return strconv.FormatInt(int64(input_num), 10)
}

func loadDataSet(filename string) ([]SliceData, error) {
	dict := make([]SliceData, 1)

	f, err := os.Open(filename)
	if err != nil {
		fmt.Println("error opening file ", err)
		return dict, err
	}
	defer f.Close()
	r := bufio.NewReader(f)
	for {
		s, err := r.ReadString('\n')
		if err == io.EOF {
			// do something here
			break
		} else if err != nil {
			return dict, err // if you return error
		} else {
			s = s[0 : len(s)-1] // remove '\n'
			s = strings.ToLower(s)
			word := strings.Split(s, "\t")

			x := SliceData{}
			if len(word) > 1 {
				x.i = stringToInt(word[0])
				x.s = word[1]
			} else {
				x.s = word[0]
			}

			dict = append(dict, x)
		}
	}

	return dict, err
}

func tokenize(sentence string) []float64 {
	// TODO: should be moved to a file
	dict := map[string]int{
		"crappy":        0,
		"ugly":          1,
		"able":          2,
		"absolute":      3,
		"amazing":       4,
		"appropriate":   5,
		"awesome":       6,
		"bad":           7,
		"beautiful":     8,
		"benefit":       9,
		"best":          10,
		"better":        11,
		"blowing":       12,
		"cheap":         13,
		"classic":       14,
		"clear":         15,
		"compact":       16,
		"compare":       17,
		"daunting":      18,
		"decent":        19,
		"definitely":    20,
		"disappoint":    21,
		"disappointed":  22,
		"disappointing": 23,
		"enjoy":         24,
		"epic":          25,
		"error":         26,
		"even":          27,
		"ever":          28,
		"every":         29,
		"excellent":     30,
		"exciting":      31,
		"extra":         32,
		"far":           33,
		"favorite":      34,
		"feel":          35,
		"genuine":       36,
		"good":          37,
		"grand":         38,
		"great":         39,
		"greatest":      40,
		"happy":         41,
		"harmful":       42,
		"hate":          43,
		"here":          44,
		"highly":        45,
		"honest":        46,
		"illogical":     47,
		"inexpensive":   48,
		"interested":    49,
		"like":          50,
		"lot":           51,
		"lots":          52,
		"love":          53,
		"lovely":        54,
		"loving":        55,
		"main":          56,
		"masterpiece":   57,
		"mind":          58,
		"mindblowing":   59,
		"misleading":    60,
		"more":          61,
		"most":          62,
		"much":          63,
		"must":          64,
		"never":         65,
		"no":            66,
		"not":           67,
		"obvious":       68,
		"perfectly":     69,
		"point":         70,
		"pretty":        71,
		"quality":       72,
		"quite":         73,
		"really":        74,
		"recommended":   75,
		"reject":        76,
		"rejected":      77,
		"respect":       78,
		"scam":          79,
		"scary":         80,
		"simple":        81,
		"simply":        82,
		"stars":         83,
		"strong":        84,
		"succinct":      85,
		"suggest":       86,
		"sure":          87,
		"terrible":      88,
		"there":         89,
		"thoroughly":    90,
		"tired":         91,
		"total":         92,
		"unimportant":   93,
		"useful":        94,
		"very":          95,
		"visit":         96,
		"well":          97,
		"winning":       98,
		"worse":         99,
		"worst":         100,
		"worth":         101,
		"worthwhile":    102,
	}

	var vec []float64
	vec = make([]float64, 102)
	sentence = strings.ToLower(sentence)
	words := myExp.Split(sentence, -1)

	for _, w := range words {
		if val, ok := dict[w]; ok {
			vec[val] = 1
		}
	}

	return vec
}

func main() {
	dict, err := loadDataSet("trainingset.txt")
	if err != nil {
		log.Fatal(err)
	}

	problem := gosvm.NewProblem()

	// We will use the words as our features
	for _, val := range dict {
		problem.Add(gosvm.TrainingInstance{float64(val.i), gosvm.FromDenseVector(tokenize(val.s))})
	}

	param := gosvm.DefaultParameters()
	param.Kernel = gosvm.NewRBFKernel(0.1) //NewPolynomialKernel(1.0, 0.1, 1)
	model, err := gosvm.TrainModel(param, problem)
	if err != nil {
		log.Fatal(err)
	}

	label1 := model.Predict(gosvm.FromDenseVector(tokenize("This is a beautiful book")))
	fmt.Printf("Predicted label: %f\n", label1)

	label2 := model.Predict(gosvm.FromDenseVector(tokenize("I hat the hot weather today..")))
	fmt.Printf("Predicted label: %f\n", label2)
}
