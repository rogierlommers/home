package quicknote

import (
	"testing"
)

func TestExtractSubjectAndBody(t *testing.T) {
	wantedSubject := "line 1"
	wantedBody := "line 2\nline 3"

	multiLine := `line 1
line 2
line 3`

	subject, body := extractSubjectAndBody(multiLine)

	if subject != wantedSubject || body != wantedBody {
		t.Fatalf("extractSubjectAndBody(): subject = %q, want %q / body = %q, want %q", subject, wantedSubject, body, wantedBody)
	}
}
