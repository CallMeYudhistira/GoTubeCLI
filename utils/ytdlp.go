package utils

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// VideoInfo holds parsed video metadata from yt-dlp
type VideoInfo struct {
	ID        string        `json:"id"`
	Title     string        `json:"title"`
	Duration  time.Duration `json:"-"`
	DurationS float64       `json:"duration"`
	Thumbnail string        `json:"thumbnail"`
	Formats   []FormatInfo  `json:"formats"`
}

// FormatInfo holds format details from yt-dlp
type FormatInfo struct {
	FormatID   string  `json:"format_id"`
	Extension  string  `json:"ext"`
	Resolution string  `json:"resolution"`
	Quality    string  `json:"format_note"`
	Filesize   int64   `json:"filesize"`
	FilesizeAp int64   `json:"filesize_approx"`
	VCodec     string  `json:"vcodec"`
	ACodec     string  `json:"acodec"`
	ABR        float64 `json:"abr"`
	VBR        float64 `json:"vbr"`
	TBR        float64 `json:"tbr"`
	FPS        float64 `json:"fps"`
	Width      int     `json:"width"`
	Height     int     `json:"height"`
}

// HasVideo returns true if this format contains a video stream
func (f *FormatInfo) HasVideo() bool {
	return f.VCodec != "" && f.VCodec != "none"
}

// HasAudio returns true if this format contains an audio stream
func (f *FormatInfo) HasAudio() bool {
	return f.ACodec != "" && f.ACodec != "none"
}

// GetFilesize returns the best available filesize estimate
func (f *FormatInfo) GetFilesize() int64 {
	if f.Filesize > 0 {
		return f.Filesize
	}
	return f.FilesizeAp
}

// ProgressCallback is called when yt-dlp reports download progress
type ProgressCallback func(percent float64, downloaded int64, total int64)

// CheckYtDlp checks if yt-dlp is installed and available in PATH
func CheckYtDlp() error {
	cmd := exec.Command("yt-dlp", "--version")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("yt-dlp is not installed or not in PATH. Please install yt-dlp: pip install yt-dlp")
	}
	return nil
}

// GetVideoInfo fetches video metadata using yt-dlp --dump-json
func GetVideoInfo(url string) (*VideoInfo, error) {
	cmd := exec.Command("yt-dlp",
		"--dump-json",
		"--no-download",
		"--no-warnings",
		url,
	)

	output, err := cmd.Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return nil, fmt.Errorf("yt-dlp failed: %s", string(exitErr.Stderr))
		}
		return nil, fmt.Errorf("failed to run yt-dlp: %w", err)
	}

	var info VideoInfo
	if err := json.Unmarshal(output, &info); err != nil {
		return nil, fmt.Errorf("failed to parse yt-dlp output: %w", err)
	}

	// Convert duration from seconds to time.Duration
	info.Duration = time.Duration(info.DurationS * float64(time.Second))

	return &info, nil
}

// DownloadWithYtDlp downloads a video/audio using yt-dlp with progress reporting
func DownloadWithYtDlp(url, outputPath string, extraArgs []string, progressCb ProgressCallback) error {
	args := []string{
		"--no-warnings",
		"-o", outputPath,
	}

	if progressCb != nil {
		// API mode: machine-parseable progress
		args = append(args,
			"--newline",
			"--progress-template", "download:[progress] %(progress._percent_str)s %(progress._downloaded_bytes_str)s %(progress._total_bytes_str)s",
		)
	}
	// CLI mode: yt-dlp shows its own beautiful progress bar by default

	args = append(args, extraArgs...)
	args = append(args, url)

	cmd := exec.Command("yt-dlp", args...)

	if progressCb != nil {
		// API mode: capture stdout and parse progress
		stderr, err := cmd.StderrPipe()
		if err != nil {
			return fmt.Errorf("failed to create stderr pipe: %w", err)
		}

		stdout, err := cmd.StdoutPipe()
		if err != nil {
			return fmt.Errorf("failed to create stdout pipe: %w", err)
		}

		if err := cmd.Start(); err != nil {
			return fmt.Errorf("failed to start yt-dlp: %w", err)
		}

		go parseYtDlpOutput(stdout, progressCb)
		stderrBytes, _ := io.ReadAll(stderr)

		if err := cmd.Wait(); err != nil {
			stderrStr := strings.TrimSpace(string(stderrBytes))
			if stderrStr != "" {
				return fmt.Errorf("yt-dlp error: %s", stderrStr)
			}
			return fmt.Errorf("yt-dlp failed: %w", err)
		}
	} else {
		// CLI mode: pass through yt-dlp output directly to terminal
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		if err := cmd.Run(); err != nil {
			return fmt.Errorf("yt-dlp failed: %w", err)
		}
	}

	return nil
}

// parseYtDlpOutput parses yt-dlp progress output and calls the callback
func parseYtDlpOutput(reader io.Reader, cb ProgressCallback) {
	if cb == nil {
		io.ReadAll(reader) // drain
		return
	}

	scanner := bufio.NewScanner(reader)
	// yt-dlp progress template outputs lines like: [progress] 45.2% 1.8MiB 4.0MiB
	progressRegex := regexp.MustCompile(`\[progress\]\s+([\d.]+)%`)

	for scanner.Scan() {
		line := scanner.Text()
		matches := progressRegex.FindStringSubmatch(line)
		if len(matches) >= 2 {
			if pct, err := strconv.ParseFloat(strings.TrimSpace(matches[1]), 64); err == nil {
				cb(pct, 0, 0)
			}
		}
	}
}

// ListFormats returns available formats for a video using yt-dlp
func ListFormats(url string) ([]FormatInfo, *VideoInfo, error) {
	info, err := GetVideoInfo(url)
	if err != nil {
		return nil, nil, err
	}
	return info.Formats, info, nil
}
