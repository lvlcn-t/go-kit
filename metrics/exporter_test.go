package metrics

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	"io"
	"io/fs"
	"testing"
	_ "unsafe"

	"github.com/google/go-cmp/cmp"
)

func Test_getTLS_Config(t *testing.T) {
	tests := []struct {
		name     string
		certPath string
		fileOpen fileOpener
		want     *tls.Config
		wantErr  bool
	}{
		{
			name:     "success",
			certPath: "testdata/cert.pem",
			fileOpen: func(name string) (fs.File, error) {
				return &mockFile{Content: pemData}, nil
			},
			want: &tls.Config{
				RootCAs:            getMockCertPool(t),
				InsecureSkipVerify: false,
				MinVersion:         tls.VersionTLS12,
			},
			wantErr: false,
		},
		{
			name:     "failure - invalid cert path",
			certPath: "testdata/invalid_cert.pem",
			fileOpen: func(name string) (fs.File, error) {
				return nil, errors.New("file not found")
			},
			wantErr: true,
		},
		{
			name:     "failure - invalid cert",
			certPath: "testdata/cert.pem",
			fileOpen: func(name string) (fs.File, error) {
				return &mockFile{Content: []byte("invalid")}, nil
			},
			wantErr: true,
		},
		{
			name:     "failure - close error",
			certPath: "testdata/cert.pem",
			fileOpen: func(name string) (fs.File, error) {
				return &mockFile{Content: pemData, CloseFunc: func() error {
					return errors.New("close error")
				}}, nil
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			openFile = tt.fileOpen
			conf, err := getTLSConfig(tt.certPath)
			if (err != nil) != tt.wantErr {
				t.Errorf("getTLSConfig() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr {
				return
			}

			// It's not possible to compare the CertPool struct directly,
			// so every field is compared separately
			if !cmp.Equal(conf.InsecureSkipVerify, tt.want.InsecureSkipVerify) {
				t.Error(cmp.Diff(conf.InsecureSkipVerify, tt.want.InsecureSkipVerify))
			}

			if !cmp.Equal(conf.MinVersion, tt.want.MinVersion) {
				t.Error(cmp.Diff(conf.MinVersion, tt.want.MinVersion))
			}

			if !cmp.Equal(conf.RootCAs, tt.want.RootCAs) {
				t.Error(cmp.Diff(conf.RootCAs, tt.want.RootCAs))
			}
		})
	}
}

func getMockCertPool(t *testing.T) *x509.CertPool {
	t.Helper()
	pool := x509.NewCertPool()
	if !pool.AppendCertsFromPEM(pemData) {
		t.Fatal("failed to append certificate(s) from file")
	}
	return pool
}

type mockFile struct {
	Content   []byte
	readPos   int
	CloseFunc func() error
}

func (mf *mockFile) Read(b []byte) (int, error) {
	if mf.readPos >= len(mf.Content) {
		return 0, io.EOF
	}
	n := copy(b, mf.Content[mf.readPos:])
	mf.readPos += n
	return n, nil
}

func (mf *mockFile) Close() error {
	if mf.CloseFunc != nil {
		return mf.CloseFunc()
	}
	return nil
}

func (mf *mockFile) Stat() (fs.FileInfo, error) {
	return nil, nil
}

var pemData = []byte(`-----BEGIN CERTIFICATE-----
MIID6TCCA1ICAQEwDQYJKoZIhvcNAQEFBQAwgYsxCzAJBgNVBAYTAlVTMRMwEQYD
VQQIEwpDYWxpZm9ybmlhMRYwFAYDVQQHEw1TYW4gRnJhbmNpc2NvMRQwEgYDVQQK
EwtHb29nbGUgSW5jLjEMMAoGA1UECxMDRW5nMQwwCgYDVQQDEwNhZ2wxHTAbBgkq
hkiG9w0BCQEWDmFnbEBnb29nbGUuY29tMB4XDTA5MDkwOTIyMDU0M1oXDTEwMDkw
OTIyMDU0M1owajELMAkGA1UEBhMCQVUxEzARBgNVBAgTClNvbWUtU3RhdGUxITAf
BgNVBAoTGEludGVybmV0IFdpZGdpdHMgUHR5IEx0ZDEjMCEGA1UEAxMaZXVyb3Bh
LnNmby5jb3JwLmdvb2dsZS5jb20wggIiMA0GCSqGSIb3DQEBAQUAA4ICDwAwggIK
AoICAQC6pgYt7/EibBDumASF+S0qvqdL/f+nouJw2T1Qc8GmXF/iiUcrsgzh/Fd8
pDhz/T96Qg9IyR4ztuc2MXrmPra+zAuSf5bevFReSqvpIt8Duv0HbDbcqs/XKPfB
uMDe+of7a9GCywvAZ4ZUJcp0thqD9fKTTjUWOBzHY1uNE4RitrhmJCrbBGXbJ249
bvgmb7jgdInH2PU7PT55hujvOoIsQW2osXBFRur4pF1wmVh4W4lTLD6pjfIMUcML
ICHEXEN73PDic8KS3EtNYCwoIld+tpIBjE1QOb1KOyuJBNW6Esw9ALZn7stWdYcE
qAwvv20egN2tEXqj7Q4/1ccyPZc3PQgC3FJ8Be2mtllM+80qf4dAaQ/fWvCtOrQ5
pnfe9juQvCo8Y0VGlFcrSys/MzSg9LJ/24jZVgzQved/Qupsp89wVidwIzjt+WdS
fyWfH0/v1aQLvu5cMYuW//C0W2nlYziL5blETntM8My2ybNARy3ICHxCBv2RNtPI
WQVm+E9/W5rwh2IJR4DHn2LHwUVmT/hHNTdBLl5Uhwr4Wc7JhE7AVqb14pVNz1lr
5jxsp//ncIwftb7mZQ3DF03Yna+jJhpzx8CQoeLT6aQCHyzmH68MrHHT4MALPyUs
Pomjn71GNTtDeWAXibjCgdL6iHACCF6Htbl0zGlG0OAK+bdn0QIDAQABMA0GCSqG
SIb3DQEBBQUAA4GBAOKnQDtqBV24vVqvesL5dnmyFpFPXBn3WdFfwD6DzEb21UVG
5krmJiu+ViipORJPGMkgoL6BjU21XI95VQbun5P8vvg8Z+FnFsvRFY3e1CCzAVQY
ZsUkLw2I7zI/dNlWdB8Xp7v+3w9sX5N3J/WuJ1KOO5m26kRlHQo7EzT3974g
-----END CERTIFICATE-----`)
