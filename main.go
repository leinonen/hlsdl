package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
)

func main() {
	baseURL := flag.String("base", "", "base URL for resolving relative segment paths (required)")
	input := flag.String("input", "", "path to .m3u8 playlist file (required)")
	outDir := flag.String("out", "segments", "output directory")
	workers := flag.Int("workers", 4, "concurrent download workers")
	referer := flag.String("referer", "", "Referer header sent with each request")
	flag.Parse()

	if *baseURL == "" || *input == "" {
		flag.Usage()
		os.Exit(1)
	}

	base, err := url.Parse(*baseURL)
	if err != nil {
		fatalf("invalid base URL: %v", err)
	}

	segments, err := parseM3U8(*input)
	if err != nil {
		fatalf("parse error: %v", err)
	}

	if err := os.MkdirAll(*outDir, 0o755); err != nil {
		fatalf("mkdir: %v", err)
	}

	fmt.Printf("downloading %d segments with %d workers\n", len(segments), *workers)

	jobs := make(chan string, len(segments))
	for _, s := range segments {
		jobs <- s
	}
	close(jobs)

	var done, failed atomic.Int64
	var wg sync.WaitGroup
	for i := 0; i < *workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for seg := range jobs {
				fullURL := resolveURL(base, seg)
				outPath := filepath.Join(*outDir, segFilename(seg))
				if err := download(fullURL, outPath, *referer); err != nil {
					fmt.Fprintf(os.Stderr, "ERROR %s: %v\n", seg, err)
					failed.Add(1)
				} else {
					n := done.Add(1)
					fmt.Printf("[%d/%d] %s\n", n, len(segments), segFilename(seg))
				}
			}
		}()
	}
	wg.Wait()

	if f := failed.Load(); f > 0 {
		fmt.Fprintf(os.Stderr, "%d segment(s) failed\n", f)
		os.Exit(1)
	}
	fmt.Println("done")
}

func parseM3U8(path string) ([]string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var segments []string
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		segments = append(segments, line)
	}
	return segments, scanner.Err()
}

func resolveURL(base *url.URL, seg string) string {
	ref, err := url.Parse(seg)
	if err != nil || ref.IsAbs() {
		return seg
	}
	return base.ResolveReference(ref).String()
}

func segFilename(seg string) string {
	// strip query string, keep only the filename part
	if idx := strings.Index(seg, "?"); idx != -1 {
		seg = seg[:idx]
	}
	return filepath.Base(seg)
}

func download(rawURL, dest, referer string) error {
	req, err := http.NewRequest(http.MethodGet, rawURL, nil)
	if err != nil {
		return err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (X11; Linux x86_64; rv:128.0) Gecko/20100101 Firefox/128.0")
	req.Header.Set("Referer", referer)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("HTTP %d", resp.StatusCode)
	}
	f, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = io.Copy(f, resp.Body)
	return err
}

func fatalf(format string, args ...any) {
	fmt.Fprintf(os.Stderr, format+"\n", args...)
	os.Exit(1)
}
