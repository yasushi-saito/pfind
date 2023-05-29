package main

import (
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/spf13/cobra"
)

type result struct {
	path string
	mode fs.FileMode
}

type pfinder struct {
	grepOutput bool
	pattern    string
	resultCh   chan result
	reqWG      sync.WaitGroup
	resultWG   sync.WaitGroup
}

func (p *pfinder) readDir(dir string) {
	defer p.reqWG.Done()

	ents, err := os.ReadDir(dir)
	if err != nil {
		log.Printf("readDir %q: %v", dir, err)
		return
	}
	for _, ent := range ents {
		matched, err := filepath.Match(p.pattern, ent.Name())
		if err != nil {
			log.Fatalf("match %q: %v", p.pattern, err)
		}
		mode := ent.Type()
		if matched {
			p.resultCh <- result{path: dir + "/" + ent.Name(), mode: mode}
		}
		if (mode & fs.ModeSymlink) != 0 {
			continue
		}
		if (mode & fs.ModeDir) != 0 {
			p.reqWG.Add(1)
			go p.readDir(dir + "/" + ent.Name())
		}
	}
}

func modeString(mode fs.FileMode) string {
	var buf strings.Builder
	if (mode & fs.ModeDir) != 0 {
		buf.WriteRune('d')
	}
	if (mode & fs.ModeSymlink) != 0 {
		buf.WriteRune('S')
	}
	return buf.String()
}

func pfind(root string, grepOutput bool, pattern string) {
	p := &pfinder{
		grepOutput: grepOutput,
		pattern:    pattern,
		resultCh:   make(chan result, 64),
	}
	p.resultWG.Add(1)
	go func() {
		defer p.resultWG.Done()
		for out := range p.resultCh {
			if p.grepOutput {
				fmt.Printf("%s:1: %s\n", out.path, modeString(out.mode))
			} else {
				fmt.Println(out.path)
			}
		}
	}()
	p.reqWG.Add(1)
	p.readDir(root)
	p.reqWG.Wait()
	close(p.resultCh)
	p.resultWG.Wait()
}

func main() {
	var (
		rootDir    string
		grepOutput bool
		rootCmd    = &cobra.Command{
			Use:   "pfind",
			Short: "Parallel find",
			Long:  "Parallel find",
			Args:  cobra.ExactArgs(1),
			Run: func(cmd *cobra.Command, args []string) {
				pfind(rootDir, grepOutput, args[0])
			},
		}
	)

	rootCmd.Flags().StringVarP(&rootDir, "dir", "d", ".", "Root directory")
	rootCmd.Flags().BoolVarP(&grepOutput, "grepoutput", "n", false, "Append a dummy linenumber to each output")
	if err := rootCmd.Execute(); err != nil {
		log.Panicf("Exec: %v", err)
	}
}
