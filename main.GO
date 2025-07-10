package main

import (
    "database/sql"
    "encoding/json"
    "fmt"
    "io"
    "log"
    "net/http"
    "os"
    "os/exec"
    "path/filepath"
    "time"

    _ "github.com/mattn/go-sqlite3"
)

 type PriceData struct {
    ASIN      string    `json:"asin"`
    Highest   float64   `json:"highest"`
    Lowest    float64   `json:"lowest"`
    Current   float64   `json:"current"`
    Average   float64   `json:"average"`
    FetchedAt time.Time `json:"fetched_at"`
}

const (
    dbFile           = "cache.db"
    recorteCacheDir  = "cache_images"
    fullCacheDir     = "full_images"
    pythonScript     = "fetch_price.py"
    jsonCacheExpiry  = time.Hour
    imageCacheExpiry = 30 * time.Minute
)

var db *sql.DB

func main() {
    if err := initCacheDirs(); err != nil {
        log.Fatal(err)
    }
    if err := initDB(); err != nil {
        log.Fatal(err)
    }
    defer db.Close()

    http.HandleFunc("/api/price", priceHandler)
    http.HandleFunc("/api/image", fullImageHandler)

    log.Println("🚀 Servidor rodando em :8080")
    log.Fatal(http.ListenAndServe(":8080", nil))
}

func initCacheDirs() error {
    for _, dir := range []string{recorteCacheDir, fullCacheDir} {
        if err := os.MkdirAll(dir, 0755); err != nil {
            return err
        }
    }
    return nil
}

func initDB() error {
    var err error
    db, err = sql.Open("sqlite3", dbFile)
    if err != nil {
        return err
    }
    _, err = db.Exec(`
        CREATE TABLE IF NOT EXISTS cache (
            asin       TEXT PRIMARY KEY,
            highest    REAL,
            lowest     REAL,
            current    REAL,
            average    REAL,
            fetched_at TEXT
        )
    `)
    return err
}

func priceHandler(w http.ResponseWriter, r *http.Request) {
    asin := r.URL.Query().Get("asin")
    if asin == "" {
        http.Error(w, "missing asin", http.StatusBadRequest)
        return
    }

    now := time.Now()
    pd, err := loadCache(asin)
    if err == nil && now.Sub(pd.FetchedAt) < jsonCacheExpiry {
        cacheFullImage(asin)
        writeJSON(w, pd)
        return
    }

    fresh, err := fetchWithPython(asin)
    if err == nil {
        fresh.FetchedAt = now
        saveCache(asin, fresh)
        cacheFullImage(asin)
        writeJSON(w, fresh)
        return
    }

    if pd != nil {
        log.Printf("⚠️ Fetch error (%v), returning cache", err)
        cacheFullImage(asin)
        writeJSON(w, pd)
        return
    }

    http.Error(w, "could not fetch price", http.StatusInternalServerError)
}

func loadCache(asin string) (*PriceData, error) {
    row := db.QueryRow(
        `SELECT highest,lowest,current,average,fetched_at FROM cache WHERE asin=?`, asin)
    var pd PriceData
    pd.ASIN = asin
    var ts string
    if err := row.Scan(&pd.Highest, &pd.Lowest, &pd.Current, &pd.Average, &ts); err != nil {
        return nil, err
    }
    pd.FetchedAt, _ = time.Parse(time.RFC3339, ts)
    return &pd, nil
}

func saveCache(asin string, pd *PriceData) {
    ts := pd.FetchedAt.Format(time.RFC3339)
    if _, err := db.Exec(`
        INSERT INTO cache(asin,highest,lowest,current,average,fetched_at)
        VALUES(?,?,?,?,?,?)
        ON CONFLICT(asin) DO UPDATE SET
            highest=excluded.highest,
            lowest=excluded.lowest,
            current=excluded.current,
            average=excluded.average,
            fetched_at=excluded.fetched_at
    `, asin, pd.Highest, pd.Lowest, pd.Current, pd.Average, ts); err != nil {
        log.Printf("❌ Error saving cache: %v", err)
    }
}

func fetchWithPython(asin string) (*PriceData, error) {
    cmd := exec.Command("python", pythonScript, "--asin", asin)
    out, err := cmd.Output()
    if err != nil {
        return nil, err
    }
    var pd PriceData
    if err := json.Unmarshal(out, &pd); err != nil {
        return nil, err
    }
    return &pd, nil
}

func cacheFullImage(asin string) {
    imgPath := filepath.Join(fullCacheDir, asin+".png")
    if info, err := os.Stat(imgPath); err == nil {
        if time.Since(info.ModTime()) < imageCacheExpiry {
            log.Printf("✅ Using cached full image for %s", asin)
            return
        }
    }

    client := &http.Client{Timeout: 15 * time.Second}
    urls := []string{
        fmt.Sprintf("https://charts.camelcamelcamel.com/us/%s/amazon.png?force=1&zero=0&w=953&h=494.5&desired=false&legend=1&ilt=1&tp=6m&fo=0&lang=en", asin),
        fmt.Sprintf("https://charts.camelcamelcamel.com/us/%s/new.png?force=1&zero=0&w=953&h=494.5&desired=false&legend=1&ilt=1&tp=6m&fo=0&lang=en", asin),
    }

    for _, url := range urls {
        log.Printf("🔍 Trying to download full image from %s", url)
        req, _ := http.NewRequest("GET", url, nil)
        req.Header.Set("User-Agent", "Mozilla/5.0")
        resp, err := client.Do(req)
        if err != nil {
            log.Printf("❌ GET error: %v", err)
            continue
        }
        if resp.StatusCode != http.StatusOK {
            log.Printf("⚠️ Status code: %d", resp.StatusCode)
            resp.Body.Close()
            continue
        }

        f, err := os.Create(imgPath)
        if err != nil {
            log.Printf("❌ File create error: %v", err)
            resp.Body.Close()
            continue
        }
        n, err := io.Copy(f, resp.Body)
        f.Close()
        resp.Body.Close()
        if err != nil {
            log.Printf("❌ Write error: %v", err)
            continue
        }
        log.Printf("✔️ Saved %d bytes to %s", n, imgPath)
        return
    }
    log.Printf("⚠️ Could not download full image for %s", asin)
}

func fullImageHandler(w http.ResponseWriter, r *http.Request) {
    asin := r.URL.Query().Get("asin")
    if asin == "" {
        http.Error(w, "missing asin", http.StatusBadRequest)
        return
    }
    imgPath := filepath.Join(fullCacheDir, asin+".png")
    http.ServeFile(w, r, imgPath)
}

func writeJSON(w http.ResponseWriter, v interface{}) {
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(v)
}
