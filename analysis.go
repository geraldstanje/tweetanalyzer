package main

import (
  "github.com/eaigner/shield"
  "log"
  "os"
  "fmt"
)

func main() {
  logger := log.New(os.Stderr, "", log.LstdFlags)
  sh := shield.New(
    shield.NewEnglishTokenizer(),
    shield.NewRedisStore("127.0.0.1:6379", "", logger, "redis"),
  )

  sh.Learn("good", "sunshine drugs love sex lobster sloth")
  sh.Learn("bad", "fear death horror government zombie god")

  c, err := sh.Classify("sloths are so cute i love them")
  if err != nil {
    fmt.Println(err.Error())
    return
  }
  if c != "good" {
    panic(c)
  }

  c, err = sh.Classify("i fear god and love the government")
  if err != nil {
    fmt.Println(err.Error())
    return
  }
  if c != "bad" {
    panic(c)
  }
}