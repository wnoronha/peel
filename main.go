package main

import (
	"embed"
	"flag"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/wnoronha/peel/internal/server"
)

var (
	Version   = "v0.1.2"
	Commit    = "dev"
	BuildDate = ""
)

//go:embed web/*
var webFS embed.FS

func main() {
	showVersion := flag.Bool("v", false, "print version and exit")
	showVersionLong := flag.Bool("version", false, "print version and exit")
	dir := flag.String("dir", "", "docs root to serve (default: ./docs if exists else cwd)")
	port := flag.Int("port", 18124, "port to listen on")
	host := flag.String("host", "127.0.0.1", "host to bind")
	flag.Parse()

	if *showVersion || *showVersionLong {
		if BuildDate != "" {
			fmt.Printf("peel %s (commit: %s, built: %s)\n", Version, Commit, BuildDate)
		} else {
			fmt.Printf("peel %s\n", Version)
		}
		os.Exit(0)
	}

	root := *dir
	if flag.NArg() > 0 && root == "" {
		root = flag.Arg(0)
	}
	if root == "" {
		if st, err := os.Stat("docs"); err == nil && st.IsDir() {
			root = "docs"
		} else {
			root = "."
		}
	}
	abs, err := filepath.Abs(root)
	if err != nil {
		fmt.Fprintln(os.Stderr, "bad dir:", err)
		os.Exit(1)
	}
	if st, err := os.Stat(abs); err != nil || !st.IsDir() {
		fmt.Fprintln(os.Stderr, "dir not found:", abs)
		os.Exit(1)
	}

	sub, err := fs.Sub(webFS, "web")
	if err != nil {
		fmt.Fprintln(os.Stderr, "failed to load embedded web assets:", err)
		os.Exit(1)
	}
	server.SetWebFS(sub)

	s := server.New(abs)
	addr := fmt.Sprintf("%s:%d", *host, *port)
	fmt.Printf("peel serving %s at http://%s\n", abs, addr)
	if err := s.Run(addr); err != nil {
		fmt.Fprintf(os.Stderr, "server error: %v\n", err)
		os.Exit(1)
	}
}
