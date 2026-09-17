package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"

	"gotube/internal/youtube"
	"gotube/utils"
)

var (
	downloadDir = "downloads"
	activeJobs  int
	maxJobs     = 3 // Limit concurrent downloads
	jobMu       sync.Mutex
)

type Server struct {
	port string
}

func NewServer(port string) *Server {
	return &Server{port: port}
}

func (s *Server) Start() error {
	utils.EnsureDir(downloadDir)

	http.Handle("/", http.FileServer(http.Dir("./public")))
	http.Handle("/downloads/", http.StripPrefix("/downloads/", http.FileServer(http.Dir(downloadDir))))

	http.HandleFunc("/api/info", s.handleInfo)
	http.HandleFunc("/api/download", s.handleDownload)
	http.HandleFunc("/api/progress", s.handleProgress)

	fmt.Printf("Starting web server on http://0.0.0.0:%s\n", s.port)
	go s.startPeriodicCleanup()
	return http.ListenAndServe("0.0.0.0:"+s.port, nil)
}

func (s *Server) handleInfo(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	url := r.URL.Query().Get("url")
	if url == "" {
		http.Error(w, "Missing url parameter", http.StatusBadRequest)
		return
	}

	svc := youtube.NewService()
	video, err := svc.GetVideo(url)
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error": "Failed to fetch video: %v"}`, err), http.StatusInternalServerError)
		return
	}

	resp := map[string]string{
		"title":     video.Title,
		"thumbnail": video.Thumbnail,
		"duration":  video.Duration.String(),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (s *Server) handleDownload(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		URL  string `json:"url"`
		Type string `json:"type"` // "video" or "audio"
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error": "Invalid JSON"}`, http.StatusBadRequest)
		return
	}

	if req.URL == "" {
		http.Error(w, `{"error": "URL is required"}`, http.StatusBadRequest)
		return
	}

	jobMu.Lock()
	if activeJobs >= maxJobs {
		jobMu.Unlock()
		http.Error(w, `{"error": "Too many active downloads. Please wait."}`, http.StatusTooManyRequests)
		return
	}
	activeJobs++
	jobMu.Unlock()

	job := Tracker.CreateJob()

	go func(j *JobInfo, requestType, url string) {
		defer func() {
			jobMu.Lock()
			activeJobs--
			jobMu.Unlock()
		}()

		svc := youtube.NewService()
		video, err := svc.GetVideo(url)
		if err != nil {
			Tracker.FailJob(j.ID, "Failed to fetch metadata")
			return
		}

		cw := &CustomWriter{JobID: j.ID}

		Tracker.UpdateProgress(j.ID, 0, StatusDownloading)

		var downloadErr error
		sanitizedTitle := utils.SanitizeFilename(video.Title)
		var finalFile string

		if requestType == "audio" {
			finalFile = sanitizedTitle + "_" + j.ID + ".mp3"
			downloadErr = svc.DownloadAudio(url, downloadDir, j.ID, true, cw)
		} else {
			finalFile = sanitizedTitle + "_" + j.ID + ".mp4"
			downloadErr = svc.DownloadVideo(url, downloadDir, "", j.ID, true, cw)
		}

		if downloadErr != nil {
			Tracker.FailJob(j.ID, downloadErr.Error())
		} else {
			Tracker.CompleteJob(j.ID, "/downloads/"+finalFile)
			// Trigger cleanup in a goroutine after 10 minutes
			go scheduleCleanup(filepath.Join(downloadDir, finalFile), 10*time.Minute)
		}
	}(job, req.Type, req.URL)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Download started",
		"id":      job.ID,
	})
}

func (s *Server) handleProgress(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	id := r.URL.Query().Get("id")
	if id == "" {
		http.Error(w, "Missing id parameter", http.StatusBadRequest)
		return
	}

	job, exists := Tracker.GetJob(id)
	if !exists {
		http.Error(w, `{"error": "Job not found"}`, http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(job)
}

func scheduleCleanup(path string, delay time.Duration) {
	time.Sleep(delay)
	os.Remove(path)
}

func (s *Server) startPeriodicCleanup() {
	ticker := time.NewTicker(1 * time.Hour)
	for range ticker.C {
		files, err := os.ReadDir(downloadDir)
		if err != nil {
			continue
		}
		for _, f := range files {
			info, err := f.Info()
			if err == nil {
				// Remove files older than 24 hours
				if time.Since(info.ModTime()) > 24*time.Hour {
					os.Remove(filepath.Join(downloadDir, f.Name()))
				}
			}
		}
	}
}
