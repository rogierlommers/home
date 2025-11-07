package quicknote

import (
	"bytes"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/rogierlommers/home/internal/mailer"
	"github.com/stretchr/testify/assert"
)

func TestHandleFileInput(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Create temp directory for testing
	tmpDir, err := os.MkdirTemp("", "quicknote_test")
	assert.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	tests := []struct {
		name          string
		filename      string
		content       []byte
		expectedError bool
	}{
		{
			name:          "valid file upload",
			filename:      "test.txt",
			content:       []byte("test content"),
			expectedError: false,
		},
		{
			name:          "missing filename header",
			filename:      "",
			content:       []byte("test content"),
			expectedError: true,
		},
		{
			name:          "empty file content",
			filename:      "empty.txt",
			content:       []byte{},
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			req := httptest.NewRequest("POST", "/", bytes.NewBuffer(tt.content))
			if tt.filename != "" {
				req.Header.Set("X-filename", tt.filename)
			}
			c.Request = req

			title, fileOnDisk, err := handleFileInput(c, tmpDir)

			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.filename, title)
				assert.Contains(t, fileOnDisk, tt.filename)

				// Verify file was created
				_, statErr := os.Stat(fileOnDisk)
				assert.NoError(t, statErr)

				// Verify file content
				fileContent, readErr := os.ReadFile(fileOnDisk)
				assert.NoError(t, readErr)
				assert.Equal(t, tt.content, fileContent)
			}
		})
	}
}

func TestDetermineTargetEmail(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		expected string
	}{
		{
			name:     "work prefix lowercase",
			content:  "w this is work content",
			expected: mailer.WorkMail,
		},
		{
			name:     "work prefix uppercase",
			content:  "W this is work content",
			expected: mailer.WorkMail,
		},
		{
			name:     "no work prefix",
			content:  "this is personal content",
			expected: mailer.PrivateMail,
		},
		{
			name:     "work prefix in middle",
			content:  "this w is not work",
			expected: mailer.PrivateMail,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := determineTargetEmail(tt.content)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestStripWorkPrefix(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		expected string
	}{
		{
			name:     "with work prefix lowercase",
			content:  "w hello world",
			expected: "hello world",
		},
		{
			name:     "with work prefix uppercase",
			content:  "W hello world",
			expected: "hello world",
		},
		{
			name:     "without work prefix",
			content:  "hello world",
			expected: "hello world",
		},
		{
			name:     "work prefix in middle",
			content:  "hello w world",
			expected: "hello w world",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := stripWorkPrefix(tt.content)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetFileType(t *testing.T) {
	tests := []struct {
		filename string
		expected string
	}{
		{"test.jpg", "JPEG Image"},
		{"test.jpeg", "JPEG Image"},
		{"test.png", "PNG Image"},
		{"test.gif", "GIF Image"},
		{"test.pdf", "PDF Document"},
		{"test.txt", "Text File"},
		{"test.doc", "Word Document"},
		{"test.docx", "Word Document"},
		{"test.zip", "ZIP Archive"},
		{"test.unknown", "UNKNOWN File"},
		{"test", "Unknown File"},
	}

	for _, tt := range tests {
		t.Run(tt.filename, func(t *testing.T) {
			result := getFileType(tt.filename)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// Benchmark tests
func BenchmarkGetFileType(b *testing.B) {
	filename := "test.jpg"
	for i := 0; i < b.N; i++ {
		getFileType(filename)
	}
}

func BenchmarkDetermineTargetEmail(b *testing.B) {
	content := "w this is work content"
	for i := 0; i < b.N; i++ {
		determineTargetEmail(content)
	}
}
