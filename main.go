package main

import (
    "fmt"
    "io"
    "net/http"
    "os"
    "os/exec"
    "runtime"
)

const (
    pdArm64URL = "https://github.com/jkawamoto/go-pixeldrain/releases/download/v0.7.6/pd_0.7.6_linux_arm64.tar.gz"
    pdAmd64URL = "https://github.com/jkawamoto/go-pixeldrain/releases/download/v0.7.6/pd_0.7.6_linux_amd64.tar.gz"
    pdBinName  = "pd"
    pdTarName  = "pd.tar.gz"
)

func getArchURL() string {
    switch runtime.GOARCH {
    case "arm64":
        return pdArm64URL
    case "amd64":
        return pdAmd64URL
    default:
        fmt.Println("Unsupported architecture:", runtime.GOARCH)
        os.Exit(1)
        return ""
    }
}

func fileExists(filename string) bool {
    info, err := os.Stat(filename)
    if os.IsNotExist(err) {
        return false
    }
    return !info.IsDir()
}

func downloadPD(url string, dest string) error {
    resp, err := http.Get(url)
    if err != nil {
        return err
    }
    defer resp.Body.Close()

    out, err := os.Create(dest)
    if err != nil {
        return err
    }
    defer out.Close()

    _, err = io.Copy(out, resp.Body)
    return err
}

func extractTarGz(src string, target string) error {
    file, err := os.Open(src)
    if err != nil {
        return err
    }
    defer file.Close()

    fmt.Println("Extracting pd binary...")
    os.Chmod(target, 0755)
    return nil
}

func pdUploadFile(pdPath string, filePath string) error {
    cmd := exec.Command(pdPath, "upload", filePath)
    cmd.Stdout = os.Stdout
    cmd.Stderr = os.Stderr
    return cmd.Run()
}

func main() {
    if len(os.Args) != 2 {
        fmt.Println("Usage: ./up <file>")
        fmt.Println("Example: ./up file.zip")
        os.Exit(1)
    }

    filePath := os.Args[1]
    
    if !fileExists(filePath) {
        fmt.Printf("File not found: %s\n", filePath)
        os.Exit(1)
    }

    if !fileExists(pdBinName) {
        archURL := getArchURL()
        fmt.Println("Downloading pd for", runtime.GOARCH)
        if err := downloadPD(archURL, pdTarName); err != nil {
            fmt.Println("Download failed:", err)
            os.Exit(1)
        }
        if err := extractTarGz(pdTarName, pdBinName); err != nil {
            fmt.Println("Extraction failed:", err)
            os.Exit(1)
        }
    } else {
        fmt.Println("pd binary found, skipping download.")
    }

    fmt.Printf("Uploading %s...\n", filePath)
    if err := pdUploadFile("./"+pdBinName, filePath); err != nil {
        fmt.Println("Upload failed:", err)
        os.Exit(1)
    }

    fmt.Println("Upload completed successfully!")
}
