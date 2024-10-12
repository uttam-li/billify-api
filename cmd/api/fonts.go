package main

import (
    "fmt"
    "io"
    "net/http"
    "os"
    "path/filepath"
)

// downloadFonts downloads the specified fonts and saves them in the ./fonts directory
func downloadFonts() error {
    // URLs of the fonts to be downloaded
    fontURLs := map[string]string{
        "Poppins-Regular.ttf": "https://github.com/google/fonts/raw/main/ofl/poppins/Poppins-Regular.ttf",
        "Poppins-Bold.ttf":    "https://github.com/google/fonts/raw/main/ofl/poppins/Poppins-Bold.ttf",
		"Poppins-SemiBold.ttf": "https://github.com/google/fonts/raw/main/ofl/poppins/Poppins-SemiBold.ttf",
        "Poppins-Italic.ttf":  "https://github.com/google/fonts/raw/main/ofl/poppins/Poppins-Italic.ttf",
    }

    // Directory to store the fonts
    fontDir := "./fonts"

    // Create the directory if it doesn't exist
    err := os.MkdirAll(fontDir, os.ModePerm)
    if err != nil {
        return fmt.Errorf("failed to create directory: %w", err)
    }

    // Download and save each font
    for fontName, fontURL := range fontURLs {
        fontPath := filepath.Join(fontDir, fontName)
        if _, err := os.Stat(fontPath); os.IsNotExist(err) {
            err := downloadFile(fontPath, fontURL)
            if err != nil {
                return fmt.Errorf("failed to download %s: %w", fontName, err)
            }
            fmt.Printf("Downloaded and saved %s\n", fontName)
        } else {
            fmt.Printf("%s already exists, skipping download\n", fontName)
        }
    }

    fmt.Println("All fonts downloaded and saved successfully.")
    return nil
}

// downloadFile downloads a file from the given URL and saves it to the specified path
func downloadFile(filepath string, url string) error {
    // Get the data
    resp, err := http.Get(url)
    if err != nil {
        return err
    }
    defer resp.Body.Close()

    // Create the file
    out, err := os.Create(filepath)
    if err != nil {
        return err
    }
    defer out.Close()

    // Write the body to file
    _, err = io.Copy(out, resp.Body)
    return err
}