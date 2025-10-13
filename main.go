package main

import (
    "archive/tar"
    "compress/gzip"
    "fmt"
    "io"
    "net/http"
    "os"
    "os/exec"
    "runtime"
    "strings"
)

const (
    pdArm64URL  = "https://github.com/jkawamoto/go-pixeldrain/releases/download/v0.7.5/pd_0.7.5_linux_arm64.tar.gz"
    pdAmd64URL  = "https://github.com/jkawamoto/go-pixeldrain/releases/download/v0.7.5/pd_0.7.5_linux_amd64.tar.gz"
    pdBinName   = "pd"
    pdTarName   = "pd.tar.gz"
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
    gz, err := gzip.NewReader(file)
    if err != nil {
        return err
    }
    defer gz.Close()
    tr := tar.NewReader(gz)
    for {
        header, err := tr.Next()
        if err == io.EOF {
            break
        }
        if err != nil {
            return err
        }
        if header.Typeflag == tar.TypeReg && strings.HasSuffix(header.Name, pdBinName) {
            out, err := os.Create(target)
            if err != nil {
                return err
            }
            defer out.Close()
            _, err = io.Copy(out, tr)
            if err != nil {
                return err
            }
            os.Chmod(target, 0755)
            break
        }
    }
    return nil
}

func pdUpload(pdPath, filePath string) error {
    cmd := exec.Command(pdPath, "upload", filePath)
    cmd.Stdout = os.Stdout
    cmd.Stderr = os.Stderr
    return cmd.Run()
}

func main() {
    if len(os.Args) != 2 {
        fmt.Println("Usage: pd <file>")
        os.Exit(1)
    }

    if fileExists(pdBinName) {
        fmt.Println("pd binary found, skipping download.")
    } else {
        archURL := getArchURL()
        fmt.Println("Downloading pd for architecture:", runtime.GOARCH)
        if err := downloadPD(archURL, pdTarName); err != nil {
            fmt.Println("Download failed:", err)
            os.Exit(1)
        }
        fmt.Println("Extracting pd binary...")
        if err := extractTarGz(pdTarName, pdBinName); err != nil {
            fmt.Println("Extraction failed:", err)
            os.Exit(1)
        }
    }

    file := os.Args[1]
    fmt.Println("Uploading file:", file)
    if err := pdUpload("./"+pdBinName, file); err != nil {
        fmt.Println("Upload failed:", err)
        os.Exit(1)
    }
    fmt.Println("Upload completed successfully!")
}
