package youtube

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"gotube/utils"
)

type Service struct{}

func NewService() *Service {
	return &Service{}
}

// VideoMeta holds essential video metadata for display and download
type VideoMeta struct {
	ID        string
	Title     string
	Duration  time.Duration
	Thumbnail string
}

// GetVideo fetches video metadata using yt-dlp
func (s *Service) GetVideo(url string) (*VideoMeta, error) {
	info, err := utils.GetVideoInfo(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch video metadata: %w", err)
	}

	return &VideoMeta{
		ID:        info.ID,
		Title:     info.Title,
		Duration:  info.Duration,
		Thumbnail: info.Thumbnail,
	}, nil
}

// PrintVideoInfo prints the video title, duration, and selected quality
func (s *Service) PrintVideoInfo(title string, duration time.Duration, quality string) {
	fmt.Printf("Title: %s\n", title)
	fmt.Printf("Duration: %s\n", formatDuration(duration))
	if quality != "" {
		fmt.Printf("Quality: %s\n", quality)
	}
	fmt.Println(strings.Repeat("-", 40))
}

// DownloadVideo handles downloading video using yt-dlp
func (s *Service) DownloadVideo(url, outputDir, quality, jobID string, force bool, customProgress io.Writer) error {
	info, err := utils.GetVideoInfo(url)
	if err != nil {
		return fmt.Errorf("failed to fetch video metadata: %w", err)
	}

	sanitizedTitle := utils.SanitizeFilename(info.Title)
	if jobID != "" {
		sanitizedTitle = fmt.Sprintf("%s_%s", sanitizedTitle, jobID)
	}
	finalOutputPath := filepath.Join(outputDir, sanitizedTitle+".mp4")

	if exists, err := utils.PromptOverwrite(finalOutputPath, force); err != nil {
		return err
	} else if !exists {
		fmt.Println("Download cancelled.")
		return nil
	}

	// Determine quality label for display
	qualityLabel := "Best Available"
	if quality != "" {
		qualityLabel = quality
	}

	s.PrintVideoInfo(info.Title, info.Duration, qualityLabel)

	// Build yt-dlp arguments
	args := []string{
		"--merge-output-format", "mp4",
	}

	if quality != "" {
		// Use bestvideo with height closest to requested quality + best audio
		args = append(args, "-f", fmt.Sprintf("bestvideo[height<=%s]+bestaudio/best[height<=%s]/best", quality, quality))
	} else {
		// Best overall quality
		args = append(args, "-f", "bestvideo+bestaudio/best")
	}

	// Set up progress reporting
	var progressCb utils.ProgressCallback
	if customProgress != nil {
		progressCb = makeWriterProgressCallback(customProgress)
	}
	// If customProgress is nil (CLI mode), yt-dlp will show its own progress bar

	fmt.Println("Downloading video...")
	if err := utils.DownloadWithYtDlp(url, finalOutputPath, args, progressCb); err != nil {
		return err
	}

	fmt.Println("Download completed successfully!")
	return nil
}

// DownloadAudio handles downloading the best audio and converting it to MP3 using yt-dlp
func (s *Service) DownloadAudio(url, outputDir, jobID string, force bool, customProgress io.Writer) error {
	info, err := utils.GetVideoInfo(url)
	if err != nil {
		return fmt.Errorf("failed to fetch video metadata: %w", err)
	}

	sanitizedTitle := utils.SanitizeFilename(info.Title)
	if jobID != "" {
		sanitizedTitle = fmt.Sprintf("%s_%s", sanitizedTitle, jobID)
	}
	finalOutputPath := filepath.Join(outputDir, sanitizedTitle+".mp3")

	if exists, err := utils.PromptOverwrite(finalOutputPath, force); err != nil {
		return err
	} else if !exists {
		fmt.Println("Download cancelled.")
		return nil
	}

	s.PrintVideoInfo(info.Title, info.Duration, "Audio Only")

	// yt-dlp arguments for audio extraction
	args := []string{
		"-x",                   // Extract audio
		"--audio-format", "mp3", // Convert to MP3
		"--audio-quality", "0", // Best quality
	}

	// Set up progress reporting
	var progressCb utils.ProgressCallback
	if customProgress != nil {
		progressCb = makeWriterProgressCallback(customProgress)
	}

	fmt.Println("Downloading audio stream...")
	if err := utils.DownloadWithYtDlp(url, finalOutputPath, args, progressCb); err != nil {
		return err
	}

	fmt.Println("Download and conversion completed successfully!")
	return nil
}

// PrintFormats prints available formats in a table using yt-dlp
func (s *Service) PrintFormats(url string) error {
	formats, info, err := utils.ListFormats(url)
	if err != nil {
		return fmt.Errorf("failed to fetch video metadata: %w", err)
	}

	s.PrintVideoInfo(info.Title, info.Duration, "")

	fmt.Printf("%-12s | %-10s | %-15s | %-10s | %s\n", "FORMAT ID", "QUALITY", "AUDIO/VIDEO", "SIZE", "EXT")
	fmt.Println(strings.Repeat("-", 70))

	for _, f := range formats {
		quality := f.Resolution
		if quality == "" || quality == "audio only" {
			quality = "Audio"
		}

		avType := "Video+Audio"
		if !f.HasVideo() && f.HasAudio() {
			avType = "Audio Only"
		} else if f.HasVideo() && !f.HasAudio() {
			avType = "Video Only"
		} else if !f.HasVideo() && !f.HasAudio() {
			continue // Skip storyboard/metadata formats
		}

		size := "N/A"
		if fs := f.GetFilesize(); fs > 0 {
			size = formatSize(fs)
		}

		fmt.Printf("%-12s | %-10s | %-15s | %-10s | %s\n", f.FormatID, quality, avType, size, f.Extension)
	}

	return nil
}

// makeWriterProgressCallback creates a ProgressCallback that updates a custom io.Writer (for API tracker)
func makeWriterProgressCallback(w io.Writer) utils.ProgressCallback {
	return func(percent float64, downloaded int64, total int64) {
		if st, ok := w.(interface{ SetTotal(int64) }); ok && total > 0 {
			st.SetTotal(total)
		}
		// Directly set progress percentage if the writer supports it
		if pw, ok := w.(interface {
			SetProgress(float64)
		}); ok {
			pw.SetProgress(percent)
		}
	}
}

// helper: format time.Duration as "4m6s"
func formatDuration(d time.Duration) string {
	d = d.Round(time.Second)
	m := int(d.Minutes())
	s := int(d.Seconds()) % 60
	if m > 0 {
		return fmt.Sprintf("%dm%ds", m, s)
	}
	return fmt.Sprintf("%ds", s)
}

// helper: format bytes into human-readable size
func formatSize(bytes int64) string {
	const (
		KB = 1024
		MB = KB * 1024
		GB = MB * 1024
	)

	switch {
	case bytes >= GB:
		return fmt.Sprintf("%.1f GB", float64(bytes)/float64(GB))
	case bytes >= MB:
		return fmt.Sprintf("%.1f MB", float64(bytes)/float64(MB))
	case bytes >= KB:
		return fmt.Sprintf("%.1f KB", float64(bytes)/float64(KB))
	default:
		return fmt.Sprintf("%d B", bytes)
	}
}

// CleanupTempFiles removes any yt-dlp temporary files in the output directory
func CleanupTempFiles(outputDir string) {
	// yt-dlp sometimes leaves .part files
	matches, _ := filepath.Glob(filepath.Join(outputDir, "*.part"))
	for _, m := range matches {
		os.Remove(m)
	}
}
