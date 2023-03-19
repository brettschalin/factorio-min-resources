package main

import (
	"io"
	"log"
	"os"

	"github.com/brettschalin/factorio-min-resources/tasscript"
)

// Usage: fmin [input file] [output file]
// This compiles a TAS-Script file into Lua code usable by a TAS running
// mod. Defaults for input/output files if not given are stdin and stdout

func main() {

	var (
		inFile  io.ReadCloser
		outFile io.WriteCloser
		err     error
	)

	if len(os.Args) > 1 {
		i := os.Args[1]
		inFile, err = os.Open(i)
		if err != nil {
			log.Fatalf("Could not open input file %q; %v", i, err)
		}
		defer inFile.Close()
	} else {
		inFile = os.Stdin
	}

	if len(os.Args) > 2 {
		of := os.Args[2]
		outFile, err = os.OpenFile(of, os.O_CREATE|os.O_RDWR|os.O_TRUNC, os.ModePerm)
		if err != nil {
			log.Fatalf(`Could not open out file %q`, of)
		}
		defer outFile.Close()
	} else {
		outFile = os.Stdout
	}

	s, err := tasscript.Read(inFile)
	if err != nil {
		log.Fatalf(`Could not parse file %q: %v`, inFile, err)
	}

	prog, err := s.ToProg()
	if err != nil {
		log.Fatalf(`Could not process file %q: %v`, inFile, err)
	}

	if err = prog.Write(outFile); err != nil {
		log.Fatalf(`Error writing output file: %v`, err)
	}
}
