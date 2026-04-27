package youtube

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	yt "github.com/kkdai/youtube/v2"
	"gotube/utils"
)

type Service struct {
	client *yt.Client
}

func NewService() *Service {
	return &Service{
		client: &yt.Client{},
	}
}

func (s *Service) GetVideo(url string) (*yt.Video, error) {
	return s.client.GetVideo(url)
}

// PrintVideoInfo prints the video title, duration, and selected quality
func (s *Service) PrintVideoInfo(video *yt.Video, quality string) {
	fmt.Printf("Title: %s\n", video.Title)
	fmt.Printf("Duration: %s\n", video.Duration.String())
	if quality != "" {
		fmt.Printf("Quality: %s\n", quality)
	}
	fmt.Println(strings.Repeat("-", 40))
}

// DownloadVideo handles downloading video and audio and merging them if necessary
func (s *Service) DownloadVideo(url, outputDir, quality string, force bool) error {
	video, err := s.GetVideo(url)
	if err != nil {
		return fmt.Errorf("failed to fetch video metadata: %w", err)
	}

	sanitizedTitle := utils.SanitizeFilename(video.Title)
	finalOutputPath := filepath.Join(outputDir, sanitizedTitle+".mp4")

	if exists, err := utils.PromptOverwrite(finalOutputPath, force); err != nil {
		return err
	} else if !exists {
		fmt.Println("Download cancelled.")
		return nil
	}

	// Determine best video and audio formats
	var videoFormat, audioFormat *yt.Format
	
	// If the user specifies a quality, try to find the closest matching video format
	if quality != "" {
		videoFormat = s.findClosestVideoFormat(video.Formats, quality)
	} else {
		// Find best overall video (usually DASH video-only has highest quality)
		videoFormat = s.getBestVideoFormat(video.Formats)
	}

	if videoFormat == nil {
		return fmt.Errorf("no suitable video format found")
	}

	// If the selected video format does not have audio, find the best audio format
	needsMerge := videoFormat.AudioChannels == 0
	if needsMerge {
		audioFormat = s.getBestAudioFormat(video.Formats)
		if audioFormat == nil {
			// Fallback: try to find a format with both
			videoFormat = s.getBestCombinedFormat(video.Formats)
			needsMerge = false
			if videoFormat == nil {
				return fmt.Errorf("no suitable video/audio format found")
			}
		}
	}

	s.PrintVideoInfo(video, videoFormat.QualityLabel)

	if !needsMerge {
		fmt.Println("Downloading combined video and audio stream...")
		return s.downloadStream(video, videoFormat, finalOutputPath, "Downloading")
	}

	// We need to download video and audio separately and merge
	tempVideoPath := filepath.Join(outputDir, "temp_video_"+sanitizedTitle+".mp4")
	tempAudioPath := filepath.Join(outputDir, "temp_audio_"+sanitizedTitle+".m4a")

	fmt.Println("Downloading video stream...")
	if err := s.downloadStream(video, videoFormat, tempVideoPath, "Video"); err != nil {
		return err
	}

	fmt.Println("Downloading audio stream...")
	if err := s.downloadStream(video, audioFormat, tempAudioPath, "Audio"); err != nil {
		return err
	}

	fmt.Println("Merging video and audio with FFmpeg...")
	if err := utils.MergeVideoAudio(tempVideoPath, tempAudioPath, finalOutputPath); err != nil {
		return fmt.Errorf("merge failed: %w", err)
	}

	fmt.Println("Download and merge completed successfully!")
	return nil
}

// DownloadAudio handles downloading the best audio and converting it to MP3
func (s *Service) DownloadAudio(url, outputDir string, force bool) error {
	video, err := s.GetVideo(url)
	if err != nil {
		return fmt.Errorf("failed to fetch video metadata: %w", err)
	}

	sanitizedTitle := utils.SanitizeFilename(video.Title)
	finalOutputPath := filepath.Join(outputDir, sanitizedTitle+".mp3")

	if exists, err := utils.PromptOverwrite(finalOutputPath, force); err != nil {
		return err
	} else if !exists {
		fmt.Println("Download cancelled.")
		return nil
	}

	audioFormat := s.getBestAudioFormat(video.Formats)
	if audioFormat == nil {
		return fmt.Errorf("no suitable audio format found")
	}

	s.PrintVideoInfo(video, "Audio Only")

	tempAudioPath := filepath.Join(outputDir, "temp_audio_"+sanitizedTitle+".m4a")
	fmt.Println("Downloading audio stream...")
	if err := s.downloadStream(video, audioFormat, tempAudioPath, "Audio"); err != nil {
		return err
	}

	fmt.Println("Converting to MP3 with FFmpeg...")
	if err := utils.ConvertToMP3(tempAudioPath, finalOutputPath); err != nil {
		return fmt.Errorf("conversion failed: %w", err)
	}

	fmt.Println("Download and conversion completed successfully!")
	return nil
}

func (s *Service) downloadStream(video *yt.Video, format *yt.Format, outputPath, progressDesc string) error {
	stream, size, err := s.client.GetStream(video, format)
	if err != nil {
		return fmt.Errorf("failed to get stream: %w", err)
	}
	defer stream.Close()

	file, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer file.Close()

	bar := utils.NewProgressBar(size, progressDesc)
	
	// Use io.MultiWriter to write to both the file and the progress bar
	writer := io.MultiWriter(file, bar)

	// Use io.CopyBuffer with a 1MB buffer to significantly improve download speed
	buf := make([]byte, 1024*1024)
	_, err = io.CopyBuffer(writer, stream, buf)
	if err != nil {
		return fmt.Errorf("failed to download stream: %w", err)
	}

	fmt.Println() // Print newline after progress bar finishes
	return nil
}

// Helper methods for format selection

func (s *Service) getBestVideoFormat(formats yt.FormatList) *yt.Format {
	// Sort by resolution
	formats.Sort()
	if len(formats) > 0 {
		return &formats[0]
	}
	return nil
}

func (s *Service) getBestAudioFormat(formats yt.FormatList) *yt.Format {
	audioFormats := formats.Type("audio")
	audioFormats.Sort() // sorts by bitrate usually
	if len(audioFormats) > 0 {
		return &audioFormats[0]
	}
	return nil
}

func (s *Service) getBestCombinedFormat(formats yt.FormatList) *yt.Format {
	combined := formats.WithAudioChannels()
	combined.Sort()
	if len(combined) > 0 {
		return &combined[0]
	}
	return nil
}

func (s *Service) findClosestVideoFormat(formats yt.FormatList, targetQuality string) *yt.Format {
	targetRes := 0
	fmt.Sscanf(targetQuality, "%dp", &targetRes)
	if targetRes == 0 {
		fmt.Sscanf(targetQuality, "%d", &targetRes)
	}

	if targetRes == 0 {
		return s.getBestVideoFormat(formats)
	}

	var bestMatch *yt.Format
	minDiff := 9999

	for i := range formats {
		f := &formats[i]
		if f.QualityLabel == "" {
			continue
		}
		
		res := 0
		fmt.Sscanf(f.QualityLabel, "%dp", &res)
		if res == 0 {
			continue
		}

		diff := res - targetRes
		if diff < 0 {
			diff = -diff
		}

		if diff < minDiff {
			minDiff = diff
			bestMatch = f
		} else if diff == minDiff && bestMatch != nil {
			// Tie breaker: higher bitrate
			if f.Bitrate > bestMatch.Bitrate {
				bestMatch = f
			}
		}
	}

	return bestMatch
}

// PrintFormats prints available formats in a table
func (s *Service) PrintFormats(url string) error {
	video, err := s.GetVideo(url)
	if err != nil {
		return fmt.Errorf("failed to fetch video metadata: %w", err)
	}

	s.PrintVideoInfo(video, "")

	fmt.Printf("%-10s | %-10s | %-15s | %-15s | %s\n", "ITAG", "QUALITY", "AUDIO/VIDEO", "BITRATE (kbps)", "MIME TYPE")
	fmt.Println(strings.Repeat("-", 80))

	// Sort formats to show highest quality first
	formats := video.Formats
	formats.Sort()

	for _, f := range formats {
		quality := f.QualityLabel
		if quality == "" {
			quality = "Audio"
		}
		
		avType := "Video+Audio"
		if f.AudioChannels == 0 {
			avType = "Video Only"
		} else if !strings.Contains(f.MimeType, "video") {
			avType = "Audio Only"
		}

		bitrate := strconv.Itoa(f.Bitrate / 1024)

		// Shorten MimeType for display
		mime := strings.Split(f.MimeType, ";")[0]

		fmt.Printf("%-10d | %-10s | %-15s | %-15s | %s\n", f.ItagNo, quality, avType, bitrate, mime)
	}

	return nil
}
