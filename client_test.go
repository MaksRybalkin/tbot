package tbot_test

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/yanzay/tbot"
)

const token = "TOKEN"

func TestNewClient(t *testing.T) {
	c := tbot.NewClient(token, nil, "https://example.com")
	if c == nil {
		t.Fatalf("client is nil")
	}
}

func TestGetMe(t *testing.T) {
	c := testClient(t, `
		{
			"ok": true,
			"result": {"id": 1}
		}
	`)

	me, err := c.GetMe()
	if err != nil {
		t.Fatalf("error on getMe: %v", err)
	}
	if me.ID == 0 {
		t.Fatalf("empty me.ID")
	}
}

func TestSendMessage(t *testing.T) {
	c := testClient(t, `
		{
			"result": {
				"chat": {"id": 1},
				"text": "helo"
			},
			"ok": true
		}
	`)

	msg, err := c.SendMessage("123", "helo")
	if err != nil {
		t.Fatalf("error on sendMessage: %v", err)
	}
	if msg.Text == "" {
		t.Fatalf("empty message text")
	}
}

func TestSendMessageWithOptions(t *testing.T) {
	c := testClient(t, `
		{
			"result": {
				"chat": {"id": 1},
				"text": "helo"
			},
			"ok": true
		}
	`)

	msg, err := c.SendMessage("123", "helo", tbot.OptParseModeMarkdown,
		tbot.OptDisableWebPagePreview, tbot.OptDisableNotification,
		tbot.OptReplyToMessageID(1), tbot.OptForceReply, tbot.OptReplyKeyboardRemove)
	if err != nil {
		t.Fatalf("error on sendMessage: %v", err)
	}
	if msg.Text == "" {
		t.Fatalf("empty message text")
	}
}

func TestForwardMessage(t *testing.T) {
	c := testClient(t, `
		{
			"ok": true,
			"result": {"message_id": 321}
		}
	`)

	msg, err := c.ForwardMessage("321", "123", 1)
	if err != nil {
		t.Fatalf("error on forwardMessage: %v", err)
	}
	if msg.MessageID == 0 {
		t.Fatalf("empty message id")
	}
}

func TestSendAudio(t *testing.T) {
	c := testClient(t, `
		{
			"ok": true,
			"result": {"message_id": 321}
		}
	`)
	msg, err := c.SendAudio("123", "aaa")
	if err != nil {
		t.Fatalf("error on sendAudio: %v", err)
	}
	if msg.MessageID == 0 {
		t.Fatalf("empty message id")
	}
}

func TestSendAudioFile(t *testing.T) {
	c := testClient(t, `
		{
			"ok": true,
			"result": {"message_id": 321}
		}
	`)
	msg, err := c.SendAudioFile("123", "client_test.go")
	if err != nil {
		t.Fatalf("error on sendAudioFile: %v", err)
	}
	if msg.MessageID == 0 {
		t.Fatalf("empty message id")
	}
}

func TestSendPhoto(t *testing.T) {
	c := testClient(t, `
		{
			"ok": true,
			"result": {"message_id": 321}
		}
	`)
	msg, err := c.SendPhoto("123", "aaa")
	if err != nil {
		t.Fatalf("error on sendPhoto: %v", err)
	}
	if msg.MessageID == 0 {
		t.Fatalf("empty message id")
	}
}

func TestSendPhotoFile(t *testing.T) {
	c := testClient(t, `
		{
			"ok": true,
			"result": {"message_id": 321}
		}
	`)
	msg, err := c.SendPhotoFile("123", "client_test.go")
	if err != nil {
		t.Fatalf("error on sendPhotoFile: %v", err)
	}
	if msg.MessageID == 0 {
		t.Fatalf("empty message id")
	}
}

func testClient(t *testing.T, resp string) *tbot.Client {
	t.Helper()
	handler := func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, resp)
	}
	httpServer := httptest.NewServer(http.HandlerFunc(handler))
	httpClient := httpServer.Client()
	return tbot.NewClient(token, httpClient, httpServer.URL)
}

func TestClient_DownloadFile(t *testing.T) {

	mux := http.NewServeMux()
	path := fmt.Sprintf("/file/bot%s/%s", token, "src/client_test.go")
	mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./client_test.go")
	})
	ts := httptest.NewServer(mux)
	httpClient := ts.Client()
	botClient := tbot.NewClient(token, httpClient, ts.URL)

	type checkFn func(*testing.T, io.Reader, error)

	wantErr := func(t *testing.T, r io.Reader, err error) {
		t.Helper()
		if err == nil {
			t.Fatal("got nil; want error")
		}
	}

	wantNilErr := func(t *testing.T, r io.Reader, err error) {
		t.Helper()
		if err != nil {
			t.Fatalf("got %v; want nil", err)
		}
	}

	cmpWithOriginal := func(wantFileName string) func(*testing.T, io.Reader, error) {
		original, err := ioutil.ReadFile(wantFileName)
		if err != nil {
			t.Errorf("failed to read original file %v", err)
		}

		return func(t *testing.T, got io.Reader, err error) {
			gotBytes, _ := ioutil.ReadAll(got)
			if !bytes.Equal(original, gotBytes) {
				t.Errorf("downloaded file and original are different")
			}
		}
	}

	testCases := []struct {
		name   string
		file   tbot.File
		checks []checkFn
	}{
		{
			name: "empty file path",
			file: tbot.File{
				FilePath: "",
			},
			checks: []checkFn{wantErr},
		},
		{
			name: "file not found",
			file: tbot.File{
				FilePath: "src/non_existed_file.go",
			},
			checks: []checkFn{wantErr},
		},
		{
			name: "existed file",
			file: tbot.File{
				FilePath: "src/client_test.go",
			},
			checks: []checkFn{wantNilErr, cmpWithOriginal("client_test.go")},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			gotReader, err := botClient.DownloadFile(tc.file)
			for _, check := range tc.checks {
				check(t, gotReader, err)
			}
		})
	}

	ts.Close()
}
