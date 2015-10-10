package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/kellegous/bungler/repo"
)

func typeFrom(s string) (repo.Type, error) {
	switch strings.ToLower(s) {
	case "jar":
		return repo.Jar, nil
	case "src":
		return repo.Src, nil
	case "doc":
		return repo.Doc, nil
	}
	return repo.Jar, fmt.Errorf("invalid type: %s", s)
}

func parseTypes(f string) ([]repo.Type, error) {
	ts := strings.Split(f, ",")

	all := map[string]bool{}

	for _, t := range ts {
		all[t] = true
	}

	res := make([]repo.Type, 0, len(all))
	for t := range all {
		tt, err := typeFrom(t)
		if err != nil {
			return nil, err
		}
		res = append(res, tt)
	}

	return res, nil
}

func main() {
	flagTypes := flag.String("types", "jar,src", "")
	flagDst := flag.String("dst", ".", "")
	flag.Parse()

	types, err := parseTypes(*flagTypes)
	if err != nil {
		log.Panic(err)
	}

	if _, err := os.Stat(*flagDst); err != nil {
		if err := os.MkdirAll(*flagDst, os.ModePerm); err != nil {
			log.Panic(err)
		}
	}

	var dep repo.Dep
	for _, arg := range flag.Args() {
		if err := dep.Parse(arg); err != nil {
			log.Panic(err)
		}

		if err := dep.Download(*flagDst, types); err != nil {
			log.Panic(err)
		}
	}
}
