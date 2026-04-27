# GoTube

GoTube is a fast, production-ready command-line application written in Go that allows you to download YouTube videos and audio efficiently.

## 🌟 Features

- **Download Video**: Fetch the highest quality video and audio, automatically merging them if necessary.
- **Download Audio Only**: Extract the best audio stream and convert it to MP3 format.
- **List Formats**: View all available qualities, bitrates, and mime types for a given video.
- **Quality Selection**: Specify your desired resolution (e.g., `720p`, `1080p`), and GoTube will find the closest match.
- **Smart Output Management**: Define output directories via flags or a configuration file (`~/.gotube/config.json`).
- **Interactive Prompts**: Safe overwrite handling to prevent accidental data loss.
- **Progress Bars**: Smooth, real-time download progress tracking in the terminal.

## ⚙️ Prerequisites

- **Go 1.25+** (if compiling from source)
- **FFmpeg**: Required for merging video/audio streams and converting audio to MP3.
  - Windows: `winget install ffmpeg` or download from [gyan.dev](https://www.gyan.dev/ffmpeg/builds/)
  - macOS: `brew install ffmpeg`
  - Linux: `sudo apt install ffmpeg`

> [!WARNING]
> Ensure `ffmpeg` is installed and accessible in your system's PATH. GoTube will fail to merge or extract audio if it cannot find the `ffmpeg` executable.

## 🚀 Installation

```bash
git clone https://github.com/CallMeYudhistira/GoTube.git
cd GoTube
go build -o gotube
```

*(Move the `gotube` executable to a folder in your PATH for global access).*

## 📖 Usage

### List Available Formats

See all available qualities and formats for a video:

```bash
gotube list <youtube_url>
```

### Download Video

Download the highest quality video and audio:

```bash
gotube download <youtube_url>
```

Download a specific quality (e.g., 720p):

```bash
gotube download <youtube_url> --quality=720
```

### Download Audio Only

Download the best audio stream and automatically convert it to MP3:

```bash
gotube audio <youtube_url>
```

### Advanced Options

**Set Output Directory**:
Use the `-o` or `--output` flag to specify where to save the files. If the directory doesn't exist, it will be created automatically.

```bash
gotube download <youtube_url> -o ./my_videos
```

**Persistent Configuration**:
You can set a default output directory by creating a config file at `~/.gotube/config.json`:

```json
{
  "output_dir": "C:/Users/username/Videos/GoTube"
}
```

## ⚠️ Disclaimer

This tool is provided for **educational purposes only**. Downloading copyrighted material without permission may violate the terms of service of YouTube and your local laws. Use responsibly.
