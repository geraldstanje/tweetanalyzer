package main

import "fmt"
import "strconv"
import "encoding/json"

func main() {
  s := `foo\u2019bar` // an escaped string
  fmt.Println("raw: ", s)

  var js string
  err := json.Unmarshal([]byte(`"`+s+`"`), &js)
  if err != nil {
    panic(err)
  }
  fmt.Println("json:", js)

  gs, err := strconv.Unquote(`"` + s + `"`)
  if err != nil {
    panic(err)
  }
  fmt.Println("go:  ", gs)

  s2 := `"foo\u2019bar"` // an escaped string
  fmt.Println("raw: ", s2)

  var js2 string
  err = json.Unmarshal([]byte(s2), &js2)
  if err != nil {
    panic(err)
  }
  fmt.Println("json:", js2)

  gs2, err := strconv.Unquote(s2)
  if err != nil {
    panic(err)
  }
  fmt.Println("go:  ", gs2)
}
