package a

import (
	"log"
	"os"
)

func fatalFunc() {
	log.Fatal("bad") // want "log.Fatal is forbidden outside main"
	os.Exit(1)       // want "os.Exit is forbidden outside main"
}
