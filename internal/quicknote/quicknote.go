package quicknote

import (
	"bytes"
	"embed"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/dustin/go-humanize"
	"github.com/gin-gonic/gin"
	"github.com/rogierlommers/home/internal/config"
	"github.com/rogierlommers/home/internal/mailer"
	"github.com/rogierlommers/home/internal/sqlitedb"
	"github.com/sirupsen/logrus"
)

const (
	InputTypeText = "text"
	InputTypeFile = "file"

	StatSourceWithAttachment = "quicknotes_with_attachment"
	StatSourceNoAttachment   = "quicknotes_no_attachment"

	WorkPrefix = "w "
)

var staticFS embed.FS

// AttachmentInfo represents file attachment metadata
type AttachmentInfo struct {
	Filename string
	Size     string
	Type     string
}

// TextInput represents the JSON payload for text input
type TextInput struct {
	Text string `json:"text" binding:"required"`
}

// NewQuicknote initializes quicknote routes
func NewQuicknote(router *gin.Engine, cfg config.AppConfig, m *mailer.Mailer, stats *sqlitedb.DB, staticHtmlFS embed.FS) {
	staticFS = staticHtmlFS
	router.POST("/api/notes/send", sendMailHandler(m, cfg, stats))
}

func sendMailHandler(m *mailer.Mailer, cfg config.AppConfig, stats *sqlitedb.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		inputType := c.GetHeader("x-input-type")
		logrus.Debugf("Processing input type: %s", inputType)

		var (
			title         string
			fileOnDisk    string
			hasAttachment bool
			err           error
		)

		switch inputType {
		case InputTypeText:
			title, err = handleTextInput(c)
			if err != nil {
				logrus.Errorf("Failed to handle text input: %v", err)
				c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid text input"})
				return
			}
			hasAttachment = false

		case InputTypeFile:
			title, fileOnDisk, err = handleFileInput(c, cfg.UploadTarget)
			if err != nil {
				logrus.Errorf("Failed to handle file input: %v", err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process file"})
				return
			}
			hasAttachment = true

		default:
			logrus.Errorf("Unknown input type: %s", inputType)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input type"})
			return
		}

		targetEmail := determineTargetEmail(title)
		if targetEmail == mailer.WorkMail {
			title = stripWorkPrefix(title)
		}

		// Send email
		htmlContent, err := renderEmailTemplate(title, fileOnDisk, hasAttachment)
		if err != nil {
			logrus.Errorf("Failed to render email template: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to prepare email"})
			return
		}

		var attachments []string
		if hasAttachment {
			attachments = []string{filepath.Base(fileOnDisk)}
		}

		if err := m.SendMail(title, targetEmail, htmlContent, attachments); err != nil {
			logrus.Errorf("Failed to send email: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to send email"})
			return
		}

		// Update statistics and respond
		statsSource := StatSourceNoAttachment
		responseMessage := fmt.Sprintf("Text note sent (%s)", humanize.Bytes(uint64(len(htmlContent))))

		if hasAttachment {
			statsSource = StatSourceWithAttachment
			fileSize := getFileSize(fileOnDisk)
			responseMessage = fmt.Sprintf("Note with attachment %s sent (%s)",
				filepath.Base(fileOnDisk), humanize.Bytes(fileSize))
		}

		if err := stats.IncrementEntry(statsSource); err != nil {
			logrus.Errorf("Failed to increment stats for %s: %v", statsSource, err)
		}

		c.JSON(http.StatusOK, gin.H{"message": responseMessage})
	}
}

// handleTextInput processes text-based quicknotes
func handleTextInput(c *gin.Context) (string, error) {
	var input TextInput
	if err := c.ShouldBindJSON(&input); err != nil {
		return "", fmt.Errorf("invalid JSON payload: %w", err)
	}
	return input.Text, nil
}

// handleFileInput processes file-based quicknotes
func handleFileInput(c *gin.Context, uploadTarget string) (string, string, error) {
	filename := c.GetHeader("X-filename")
	if filename == "" {
		return "", "", fmt.Errorf("missing X-filename header")
	}

	bodyBytes, err := io.ReadAll(c.Request.Body)
	if err != nil {
		return "", "", fmt.Errorf("failed to read request body: %w", err)
	}

	if len(bodyBytes) == 0 {
		return "", "", fmt.Errorf("empty file content")
	}

	fileOnDisk := filepath.Join(uploadTarget, filename)
	if err := os.WriteFile(fileOnDisk, bodyBytes, 0644); err != nil {
		return "", "", fmt.Errorf("failed to write file: %w", err)
	}

	return filename, fileOnDisk, nil
}

// renderEmailTemplate generates HTML content for the email
func renderEmailTemplate(title, fileOnDisk string, hasAttachment bool) (string, error) {
	htmlBytes, err := staticFS.ReadFile("static_html/quicknote_received.html")
	if err != nil {
		return "", fmt.Errorf("failed to read template file: %w", err)
	}

	tmpl, err := template.New("quicknote_received.html").Parse(string(htmlBytes))
	if err != nil {
		return "", fmt.Errorf("failed to parse template: %w", err)
	}

	data := struct {
		Title      string
		Message    string
		Attachment *AttachmentInfo
	}{
		Title:   title,
		Message: "Your quicknote has been received successfully.",
	}

	if hasAttachment {
		data.Attachment = createAttachmentInfo(fileOnDisk)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("failed to execute template: %w", err)
	}

	return buf.String(), nil
}

// createAttachmentInfo creates attachment metadata
func createAttachmentInfo(fileOnDisk string) *AttachmentInfo {
	fileInfo, err := os.Stat(fileOnDisk)
	if err != nil {
		logrus.Errorf("Failed to get file info for %s: %v", fileOnDisk, err)
		return &AttachmentInfo{
			Filename: filepath.Base(fileOnDisk),
			Size:     "Unknown",
			Type:     "Unknown",
		}
	}

	return &AttachmentInfo{
		Filename: filepath.Base(fileOnDisk),
		Size:     humanize.Bytes(uint64(fileInfo.Size())),
		Type:     getFileType(fileOnDisk),
	}
}

// determineTargetEmail determines email destination based on content prefix
func determineTargetEmail(content string) string {
	if strings.HasPrefix(strings.ToLower(content), WorkPrefix) {
		return mailer.WorkMail
	}
	return mailer.PrivateMail
}

// getFileSize safely gets file size
func getFileSize(filename string) uint64 {
	fileInfo, err := os.Stat(filename)
	if err != nil {
		logrus.Errorf("Failed to get file size for %s: %v", filename, err)
		return 0
	}
	return uint64(fileInfo.Size())
}

// getFileType determines file type from extension
func getFileType(filename string) string {
	ext := strings.ToLower(filepath.Ext(filename))
	switch ext {
	case ".jpg", ".jpeg":
		return "JPEG Image"
	case ".png":
		return "PNG Image"
	case ".gif":
		return "GIF Image"
	case ".pdf":
		return "PDF Document"
	case ".txt":
		return "Text File"
	case ".doc", ".docx":
		return "Word Document"
	case ".zip":
		return "ZIP Archive"
	default:
		if ext != "" {
			return strings.ToUpper(ext[1:]) + " File"
		}
		return "Unknown File"
	}
}

// stripWorkPrefix removes work prefix from content
func stripWorkPrefix(content string) string {
	if strings.HasPrefix(strings.ToLower(content), WorkPrefix) {
		return content[2:]
	}
	return content
}
