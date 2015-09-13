package main

import "encoding/hex"
import "fmt"
import "os"

import "github.com/hillsg/sha/mysha1"

func main() {
	hash := mysha1.Digest(os.Stdin)
	fmt.Println(hex.EncodeToString(hash))
}
