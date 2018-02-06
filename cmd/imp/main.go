package main

import (
	"flag"
	"fmt"
	"os"

	"jw4.us/import/pkg"
)

var (
	dbFile = "repo.db"
)

func init() {
	flag.StringVar(&dbFile, "db", dbFile, "BoltDB file name")
}

func main() {
	flag.Parse()
	store := pkg.NewStore(dbFile)
	if err := store.Initialize(nil); err != nil {
		fmt.Printf("error initializing db: %v\n", err)
		os.Exit(-1)
	}
	repos, err := store.List("")
	if err != nil {
		fmt.Printf("error listing db: %v\n", err)
		os.Exit(-1)
	}

	fmt.Printf("found %d repos\n", len(repos))
	for ix, repo := range repos {
		fmt.Printf("%3d) %s\n", ix, repo.String())
	}
}
