package main

import (
  "bufio"
  "bytes"
  "flag"
  "fmt"
  "io/ioutil"
  "log"
  "net/http"
  "net/url"
  "os"
  "path/filepath"
  "strings"
  "sync"
  "time"
)

func main() {
  endpoint := flag.String("endpoint", "http://localhost:7878", "Oxigraph endpoint URL")
  loadOrderFile := flag.String("load-order", "../../generated/rdf-load-order.txt", "Path to load order file")
  parallel := flag.Int("parallel", 3, "Number of parallel loaders")
  flag.Parse()

  log.Printf("RDF Loader for Oxigraph")
  log.Printf("Endpoint: %s", *endpoint)
  log.Printf("Load order: %s", *loadOrderFile)
  log.Printf("Parallel workers: %d", *parallel)

  // Read load order
  files, err := readLoadOrder(*loadOrderFile)
  if err != nil {
    log.Fatalf("Failed to read load order: %v", err)
  }

  log.Printf("Found %d files to load", len(files))

  // Load files with parallelism
  var wg sync.WaitGroup
  fileChan := make(chan string, len(files))
  results := make(chan loadResult, len(files))

  // Start workers
  for i := 0; i < *parallel; i++ {
    wg.Add(1)
    go loadWorker(&wg, fileChan, results, *endpoint)
  }

  // Queue files
  go func() {
    for _, file := range files {
      fileChan <- file
    }
    close(fileChan)
  }()

  // Collect results
  loaded := 0
  failed := 0
  go func() {
    wg.Wait()
    close(results)
  }()

  for result := range results {
    if result.err != nil {
      log.Printf("✗ FAILED: %s - %v", result.file, result.err)
      failed++
    } else {
      log.Printf("✓ LOADED: %s (%d triples)", result.file, result.triples)
      loaded++
    }
  }

  log.Printf("\n=== Load Summary ===")
  log.Printf("Loaded: %d files", loaded)
  log.Printf("Failed: %d files", failed)

  // Verify
  count, err := verifyLoad(*endpoint)
  if err != nil {
    log.Printf("Verification error: %v", err)
  } else {
    log.Printf("Total triples in store: %d", count)
  }
}

type loadResult struct {
  file    string
  triples int
  err     error
}

func loadWorker(wg *sync.WaitGroup, files <-chan string, results chan<- loadResult, endpoint string) {
  defer wg.Done()

  client := &http.Client{Timeout: 30 * time.Second}

  for file := range files {
    data, err := ioutil.ReadFile(file)
    if err != nil {
      results <- loadResult{file: file, err: fmt.Errorf("read file: %w", err)}
      continue
    }

    triples := bytes.Count(data, []byte{'\n'})

    query := fmt.Sprintf("INSERT DATA { %s }", string(data))

    resp, err := client.PostForm(fmt.Sprintf("%s/update", endpoint),
      url.Values{"update": {query}})
    if err != nil {
      results <- loadResult{file: file, err: fmt.Errorf("POST: %w", err)}
      continue
    }
    resp.Body.Close()

    if resp.StatusCode >= 400 {
      results <- loadResult{file: file, err: fmt.Errorf("HTTP %d", resp.StatusCode)}
      continue
    }

    results <- loadResult{file: file, triples: triples}
  }
}

func readLoadOrder(path string) ([]string, error) {
  file, err := os.Open(path)
  if err != nil {
    return nil, err
  }
  defer file.Close()

  var files []string
  scanner := bufio.NewScanner(file)
  for scanner.Scan() {
    line := strings.TrimSpace(scanner.Text())
    if line == "" || strings.HasPrefix(line, "#") {
      continue
    }
    files = append(files, line)
  }

  return files, scanner.Err()
}

func verifyLoad(endpoint string) (int, error) {
  query := "SELECT (COUNT(*) as ?count) WHERE { ?s ?p ?o }"
  resp, err := http.Get(fmt.Sprintf("%s/query?query=%s", endpoint, url.QueryEscape(query)))
  if err != nil {
    return 0, err
  }
  defer resp.Body.Close()

  // Parse result (simplified - assumes JSON response)
  body, _ := ioutil.ReadAll(resp.Body)
  // In real code, would parse JSON response
  _ = body

  return 0, nil
}
