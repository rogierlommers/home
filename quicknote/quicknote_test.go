package quicknote

import (
	"testing"
)

func TestExtractSubjectAndBodyMultiliner(t *testing.T) {
	wantedSubject := "line 1"
	wantedBody := "line 2\nline 3"

	multiLine := `line 1
line 2
line 3`

	subject, body := extractSubjectAndBody(multiLine)

	if subject != wantedSubject || body != wantedBody {
		t.Fatalf("TestExtractSubjectAndBodyMultiliner(): subject = %q, want %q / body = %q, want %q", subject, wantedSubject, body, wantedBody)
	}
}

func TestExtractSubjectAndBodySingleLiner(t *testing.T) {
	wantedSubject := "line 1"
	wantedBody := "line 1"

	sinleLine := `line 1`

	subject, body := extractSubjectAndBody(sinleLine)

	if subject != wantedSubject || body != wantedBody {
		t.Fatalf("TestExtractSubjectAndBodySingleLiner(): subject = %q, want %q / body = %q, want %q", subject, wantedSubject, body, wantedBody)
	}

}
