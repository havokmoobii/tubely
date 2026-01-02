package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"crypto/rand"
	"encoding/base64"
	"os/exec"
	"bytes"
	"encoding/json"
)

func (cfg apiConfig) ensureAssetsDir() error {
	if _, err := os.Stat(cfg.assetsRoot); os.IsNotExist(err) {
		return os.Mkdir(cfg.assetsRoot, 0755)
	}
	return nil
}

func getAssetPath(mediaType string) string {
	key := make([]byte, 32)
	rand.Read(key)
	assetID := base64.RawURLEncoding.EncodeToString(key)

	ext := mediaTypeToExt(mediaType)
	return fmt.Sprintf("%s%s", assetID, ext)
}

func (cfg apiConfig) getAssetDiskPath(assetPath string) string {
	return filepath.Join(cfg.assetsRoot, assetPath)
}

func (cfg apiConfig) getAssetURL(assetPath string) string {
	return fmt.Sprintf("http://localhost:%s/assets/%s", cfg.port, assetPath)
}

func (cfg apiConfig) getS3AssetURL(key string) string {
	return fmt.Sprintf("https://%s.s3.%s.amazonaws.com/%s", cfg.s3Bucket, cfg.s3Region, key)
}

func mediaTypeToExt(mediaType string) string {
	parts := strings.Split(mediaType, "/")
	if len(parts) != 2 {
		return ".bin"
	}
	return "." + parts[1]
}

func getVideoAspectRatio(filePath string) (string, error) {
	cmd := exec.Command("ffprobe", "-v", "error", "-print_format", "json", "-show_streams", filePath)
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		fmt.Println("Can't get video aspect ratio")
		return "", err
	}

	type ffProbeResponse struct {
		Streams []struct {
			Width              int    `json:"width,omitempty"`
			Height             int    `json:"height,omitempty"`
		} `json:"streams"`
	}

	ffProbeOut := ffProbeResponse{}

	err = json.Unmarshal(out.Bytes(), &ffProbeOut)
	if err != nil {
		fmt.Println("Can't umarshall json for ffprobe")
		return "", err
	}

	if ffProbeOut.Streams[0].Width == ffProbeOut.Streams[0].Height {
		return "other", nil
	}
	if ffProbeOut.Streams[0].Width > ffProbeOut.Streams[0].Height {
		return "16:9", nil
	}
	return "9:16", nil
}

func getVideoAssetPath(assetPath string, aspectRatio string) string {
	if aspectRatio == "16:9" {
		return fmt.Sprintf("landscape/%v", assetPath)
	}
	if aspectRatio == "9:16" {
		return fmt.Sprintf("portrait/%v", assetPath)
	}
	return fmt.Sprintf("other/%v", assetPath)
}

func processVideoForFastStart(filePath string) (string, error) {
	outPath := filePath + ".processing"
	cmd := exec.Command("ffmpeg", "-i", filePath, "-c", "copy", "-movflags", "faststart", "-f", "mp4", outPath)
	err := cmd.Run()
	if err != nil {
		fmt.Println("Can't preprocess video")
		return "", err
	}
	return outPath, nil
}