package main

import (
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"sync"

	"github.com/spf13/cobra"
)

type pfinder struct {
	pattern  string
	resultCh chan string
	reqWG    sync.WaitGroup
	resultWG sync.WaitGroup
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
		if matched {
			p.resultCh <- dir + "/" + ent.Name()
		}
		mode := ent.Type()
		if (mode & fs.ModeSymlink) != 0 {
			continue
		}
		if (mode & fs.ModeDir) != 0 {
			p.reqWG.Add(1)
			go p.readDir(dir + "/" + ent.Name())
		}
	}
}

func pfind(root, pattern string) {
	p := &pfinder{
		pattern:  pattern,
		resultCh: make(chan string, 64),
	}
	p.resultWG.Add(1)
	go func() {
		defer p.resultWG.Done()
		for out := range p.resultCh {
			fmt.Println(out)
		}
	}()
	p.reqWG.Add(1)
	p.readDir(root)
	p.reqWG.Wait()
	close(p.resultCh)
	p.resultWG.Wait()
}

func main() {
	var rootDir string
	var rootCmd = &cobra.Command{
		Use:   "pfind",
		Short: "Parallel find",
		Long:  "Parallel find",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			log.Printf("Rootdir: %v args %v", rootDir, args)
			pfind(rootDir, args[0])
		},
	}

	rootCmd.Flags().StringVarP(&rootDir, "dir", "d", ".", "Root directory")
	if err := rootCmd.Execute(); err != nil {
		log.Panicf("Exec: %v", err)
	}
}
