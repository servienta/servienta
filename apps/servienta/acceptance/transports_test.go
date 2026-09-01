package acceptance

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"net/http"
	"testing"
	"time"

	scp "github.com/bramvdbogaerde/go-scp"
	"github.com/jlaffaye/ftp"
	"github.com/pin/tftp/v3"
	"golang.org/x/crypto/ssh"
)

const (
	filesUser     = "servienta"
	filesPassword = "throwaway-not-a-secret"
)

// --- R1 all five transports, criterion 2: same file, byte-for-byte ---
func TestFixtureByteCompareAllTransports(t *testing.T) {
	e := startEngine(t)
	for name, want := range fixtureData {
		for transport, fetch := range map[string]func(*testing.T, *engine, string) []byte{
			"http":  fetchHTTP,
			"https": fetchHTTPS,
			"ftp":   fetchFTP,
			"tftp":  fetchTFTP,
			"scp":   fetchSCP,
		} {
			got := fetch(t, e, name)
			if !bytes.Equal(got, want) {
				t.Fatalf("%s: %s served %d bytes differing from original %d", transport, name, len(got), len(want))
			}
		}
	}
}

// --- R1: HTTPS, FTP, SCP reject wrong credentials ---
func TestFilesAuthRequired(t *testing.T) {
	e := startEngine(t)
	// HTTPS without credentials
	c := &http.Client{Transport: &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}}}
	res, err := c.Get("https://" + e.endpoints["files-https"] + "/small.txt")
	if err != nil {
		t.Fatal(err)
	}
	res.Body.Close()
	if res.StatusCode != http.StatusUnauthorized {
		t.Fatalf("https without creds: want 401, got %d", res.StatusCode)
	}
	// FTP with a wrong password
	conn, err := ftp.Dial(e.endpoints["files-ftp"], ftp.DialWithTimeout(3*time.Second))
	if err != nil {
		t.Fatal(err)
	}
	if err := conn.Login(filesUser, "wrong"); err == nil {
		t.Fatal("ftp accepted a wrong password")
	}
	conn.Quit()
	// SCP with a wrong password
	_, err = ssh.Dial("tcp", e.endpoints["files-scp"], &ssh.ClientConfig{
		User:            filesUser,
		Auth:            []ssh.AuthMethod{ssh.Password("wrong")},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         3 * time.Second,
	})
	if err == nil {
		t.Fatal("scp accepted a wrong password")
	}
}

func fetchHTTP(t *testing.T, e *engine, name string) []byte {
	t.Helper()
	res, err := http.Get(e.files + "/" + name)
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	var buf bytes.Buffer
	buf.ReadFrom(res.Body)
	return buf.Bytes()
}

func fetchHTTPS(t *testing.T, e *engine, name string) []byte {
	t.Helper()
	c := &http.Client{Transport: &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}}}
	req, _ := http.NewRequest("GET", "https://"+e.endpoints["files-https"]+"/"+name, nil)
	req.SetBasicAuth(filesUser, filesPassword)
	res, err := c.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		t.Fatalf("https %s: %d", name, res.StatusCode)
	}
	var buf bytes.Buffer
	buf.ReadFrom(res.Body)
	return buf.Bytes()
}

func fetchFTP(t *testing.T, e *engine, name string) []byte {
	t.Helper()
	conn, err := ftp.Dial(e.endpoints["files-ftp"], ftp.DialWithTimeout(5*time.Second))
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Quit()
	if err := conn.Login(filesUser, filesPassword); err != nil {
		t.Fatal(err)
	}
	r, err := conn.Retr(name)
	if err != nil {
		t.Fatal(err)
	}
	defer r.Close()
	var buf bytes.Buffer
	buf.ReadFrom(r)
	return buf.Bytes()
}

func fetchTFTP(t *testing.T, e *engine, name string) []byte {
	t.Helper()
	c, err := tftp.NewClient(e.endpoints["files-tftp"])
	if err != nil {
		t.Fatal(err)
	}
	c.SetBlockSize(65456) // large fixture at a sane number of round trips
	wt, err := c.Receive(name, "octet")
	if err != nil {
		t.Fatal(err)
	}
	var buf bytes.Buffer
	if _, err := wt.WriteTo(&buf); err != nil {
		t.Fatal(err)
	}
	return buf.Bytes()
}

func fetchSCP(t *testing.T, e *engine, name string) []byte {
	t.Helper()
	client := scp.NewClient(e.endpoints["files-scp"], &ssh.ClientConfig{
		User:            filesUser,
		Auth:            []ssh.AuthMethod{ssh.Password(filesPassword)},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         5 * time.Second,
	})
	if err := client.Connect(); err != nil {
		t.Fatal(err)
	}
	defer client.Close()
	var buf bytes.Buffer
	if err := client.CopyFromRemotePassThru(t.Context(), nopWriteCloser{&buf}, name, nil); err != nil {
		t.Fatal(fmt.Errorf("scp %s: %w", name, err))
	}
	return buf.Bytes()
}

type nopWriteCloser struct{ *bytes.Buffer }

func (nopWriteCloser) Close() error { return nil }
