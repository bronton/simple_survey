# Survey Voice Recorder

## Project Overview

Survey Voice Recorder is a Go-based web application that allows you to conduct surveys while simultaneously recording the respondent's voice. The collected responses and audio recording are automatically packaged into a ZIP archive and sent to a specified email address.


## Table of Contents

- [Features](#features)
- [Architecture](#architecture)
- [Installation](#installation)
- [Configuration](#configuration)
- [Running the Application](#running-the-application)
- [API](#api)
- [User Interface](#user-interface)
- [Working with Questions](#working-with-questions)
- [Audio Processing](#audio-processing)
- [Security](#security)
- [Troubleshooting](#troubleshooting)
- [Extending Functionality](#extending-functionality)
- [FAQ](#faq)

## Features

### Core Features:
- 📝 **Flexible Survey System** — Support for various question types
- 🎙️ **Voice Recording** — Synchronized audio recording from user's microphone
- 📊 **Data Export** — Saving responses in CSV format
- 📧 **Automatic Delivery** — Results sent to email as a ZIP archive
- ⚙️ **Configurability** — Configuration through JSON file

### Supported Question Types:
1. **Single Choice** (single_choice) — One answer from multiple options
2. **Multiple Choice** (multi_choice) — Several answers from a list
3. **Text Input** (text) — Free-form response
4. **Combined** (mixed) — Selection from list + "custom option"

## Architecture

The application uses a modular architecture with separation of concerns:

```
survey-voice-recorder/
├── main.go           // Entry point, initialization and server launch
├── config.go         // Configuration loading and validation
├── session.go        // User session management
├── audio.go          // Audio recording and processing
├── response.go       // User response handling
├── email.go          // Sending results via email
├── utils.go          // Helper functions
├── config.json       // Configuration file
├── templates/        // HTML templates
│   ├── survey.html   // Survey form
│   └── complete.html // Completion page
├── static/           // Static files
├── uploads/          // Temporary storage
│   └── responses/    // Directory for responses
└── go.mod            // Project dependencies
```

### Technologies Used:
- **Backend**: Go 1.16+
- **Audio**: PortAudio, WAV
- **Web**: HTML, JavaScript, CSS
- **Data**: CSV, ZIP
- **Email**: SMTP

## Installation

### Prerequisites

- Go 1.16 or higher
- Git
- PortAudio (system library)

### Installing Dependencies

#### Linux (Debian/Ubuntu)
```bash
# Installing Go (if not installed)
wget https://golang.org/dl/go1.16.linux-amd64.tar.gz
sudo tar -C /usr/local -xzf go1.16.linux-amd64.tar.gz
export PATH=$PATH:/usr/local/go/bin

# Installing PortAudio
sudo apt-get update
sudo apt-get install -y portaudio19-dev
```

#### macOS
```bash
# Installing Homebrew (if not installed)
/bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)"

# Installing Go and PortAudio
brew install go portaudio
```

#### Windows
1. Download and install Go from the [official website](https://golang.org/dl/)
2. Download the PortAudio library from the [official website](http://www.portaudio.com/download.html)
3. Add Go and PortAudio to the PATH variable

### Cloning the Repository
```bash
git clone https://github.com/yourusername/survey-voice-recorder.git
cd survey-voice-recorder
```

### Installing Go Dependencies
```bash
go mod init survey-voice-recorder
go get github.com/gordonklaus/portaudio
go get github.com/youpy/go-wav
go get github.com/jordan-wright/email
go get github.com/google/uuid
go mod tidy
```

### Creating Directory Structure
```bash
mkdir -p uploads/responses
mkdir -p static
mkdir -p templates
```

## Configuration

### Configuration File

Create a `config.json` file based on the example:

```bash
cp sample-config.json config.json
```

Edit `config.json` according to your requirements:

```json
{
  "email": {
    "to": "your.email@example.com",
    "from": "survey-system@yourdomain.com",
    "subject": "Survey Results"
  },
  "smtp_host": "smtp.yourdomain.com",
  "smtp_port": 587,
  "smtp_user": "username",
  "smtp_pass": "password",
  "questions": [
    // Question configuration...
  ]
}
```

### HTML Templates

Place templates in the `templates/` directory:

```bash
cp templates-survey.html templates/survey.html
cp templates-complete.html templates/complete.html
```

## Running the Application

### Running in Development Mode
```bash
go run *.go
```

### Building and Running Executable
```bash
go build -o survey-app
./survey-app
```

### Command Line Parameters
```bash
# Specifying port
./survey-app -port 3000

# Specifying configuration path
./survey-app -config ./custom-config.json
```

### Running with systemd (Linux)

Create a file `/etc/systemd/system/survey-app.service`:

```ini
[Unit]
Description=Survey Voice Recorder
After=network.target

[Service]
User=www-data
WorkingDirectory=/path/to/survey-voice-recorder
ExecStart=/path/to/survey-voice-recorder/survey-app
Restart=on-failure

[Install]
WantedBy=multi-user.target
```

Then activate the service:
```bash
sudo systemctl enable survey-app
sudo systemctl start survey-app
```

## API

The application provides the following HTTP endpoints:

| Path | Method | Description |
|------|-------|----------|
| `/` | GET | Redirect to survey page |
| `/survey` | GET | Survey form page |
| `/start-recording` | GET | Start audio recording (parameter: `session_id`) |
| `/stop-recording` | GET | Stop audio recording (parameter: `session_id`) |
| `/submit` | POST | Submit form with responses |
| `/complete` | GET | Completion page |
| `/static/*` | GET | Static files |

## User Interface

### Survey Page
![Survey Page](https://via.placeholder.com/600x400?text=Survey+Page)

Main interface elements:
- Survey title
- Audio recording control button
- List of questions of different types
- Submit button

### Completion Page
![Completion Page](https://via.placeholder.com/600x400?text=Completion+Page)

## Working with Questions

### Question Structure in Configuration
```json
{
  "id": "q1",
  "text": "Question text",
  "type": "single_choice|multi_choice|text|mixed",
  "options": ["Option 1", "Option 2", ...],
  "allow_custom": true,
  "required": true
}
```

### Examples of Different Question Types

#### Single Choice
```json
{
  "id": "satisfaction",
  "text": "How satisfied are you with our service?",
  "type": "single_choice",
  "options": [
    "Very satisfied",
    "Somewhat satisfied",
    "Neutral",
    "Somewhat dissatisfied",
    "Very dissatisfied"
  ],
  "required": true
}
```

#### Multiple Choice
```json
{
  "id": "features",
  "text": "Which features do you use?",
  "type": "multi_choice",
  "options": [
    "Analytics",
    "Reports",
    "Integrations",
    "API",
    "Mobile App"
  ],
  "required": false
}
```

#### Text Input
```json
{
  "id": "feedback",
  "text": "Share your feedback on how we can improve our service:",
  "type": "text",
  "required": false
}
```

#### Combined
```json
{
  "id": "source",
  "text": "How did you hear about us?",
  "type": "mixed",
  "options": [
    "Search engines",
    "Advertisement",
    "Social media",
    "From friends",
    "Conference"
  ],
  "allow_custom": true,
  "required": true
}
```

## Audio Processing

### Audio Recording Process

1. User clicks "Start Recording" button
2. Browser requests permission to access microphone
3. Client sends request to server to create recording
4. Server initiates recording through PortAudio
5. Audio data is buffered in memory
6. When the survey is completed, recording stops
7. Audio is saved in WAV format

### Recording Parameters
- Sample rate: 44100 Hz
- Channels: 1 (mono)
- Bit depth: 16 bit

### File Format
- Container: WAV
- Codec: PCM (uncompressed)
- Approximate size: ~5 MB per minute of recording

## Security

### Security Recommendations

1. **Configuration File Protection**
   - Restrict access to `config.json` (contains SMTP credentials)
   - Use environment variables for sensitive data

2. **HTTPS**
   - Use HTTPS to protect transmitted data
   - Modern browsers require HTTPS for audio recording

3. **Data Storage**
   - Files are stored temporarily and deleted after sending
   - Restrict access to the `uploads` directory

4. **Input Validation**
   - All user data undergoes validation
   - Protection against XSS attacks in templates

### Privacy Policy

It is recommended to develop and provide users with a privacy policy that explains:
- What data is collected
- How audio recording is used
- Data retention periods
- Information protection measures

## Troubleshooting

### Common Errors

#### 1. Microphone Access Error
**Symptom**: Unable to start recording, microphone access error in browser console.

**Solution**:
- Ensure the site uses HTTPS or localhost
- Check browser permissions
- Make sure the microphone is connected and working

#### 2. Email Sending Error
**Symptom**: Survey completes, but email doesn't arrive.

**Solution**:
- Check SMTP settings in configuration
- Ensure SMTP port is not blocked by firewall
- Check application logs for errors

#### 3. "portaudio: not initialized" Error
**Symptom**: Application doesn't start with portaudio error.

**Solution**:
- Ensure PortAudio library is installed
- Check permissions to audio devices
- Restart computer

## Extending Functionality

### Ideas for Further Development

1. **Administrator Authorization**
   - Creating an admin panel
   - Viewing survey statistics

2. **Audio Analytics**
   - Automatic transcription of audio recordings
   - Analysis of speech emotional tone

3. **Additional Question Types**
   - Rating scales
   - Matrix questions
   - File uploads

4. **Integrations**
   - Export data to Google Sheets
   - Integration with CRM systems
   - Webhook for survey events

### Contributing to the Project

We welcome contributions to the project! If you want to contribute:
1. Fork the repository
2. Create a branch with new functionality
3. Submit a pull request with a description of changes

## FAQ

### General Questions

**Q: How many users can take the survey simultaneously?**  
A: Theoretically there are no limits, but in practice it's recommended to have no more than 50-100 simultaneous users per server with 2GB RAM due to resource consumption during audio recording.

**Q: How long can audio recordings be stored?**  
A: Audio recordings are stored temporarily until sent via email and then deleted from the server. If you need long-term storage, it's recommended to set up email message archiving.

**Q: Which browsers are supported?**  
A: The application supports modern browsers with WebRTC:
- Chrome 49+
- Firefox 52+
- Edge 79+
- Safari 11+
- Opera 36+

**Q: Are mobile devices supported?**  
A: Yes, the application works on mobile devices with WebRTC support.

### Technical Questions

**Q: How to change the maximum recording duration?**  
A: In the current version there is no direct duration limit. Recording continues until explicitly stopped or the survey is completed.

**Q: Can I use a different audio format instead of WAV?**  
A: The current version only uses WAV. Supporting other formats would require code modification in `audio.go`.

**Q: What size ZIP archive can be sent via email?**  
A: This depends on your SMTP server limitations. Most servers have a 10-25 MB limit.
