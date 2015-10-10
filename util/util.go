package util

import (
	"bytes"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"hash"
	"io"
	"io/ioutil"
	"net/http"
	"os"
)

func getShaFor(url string) ([]byte, error) {
	res, err := http.Get(url + ".sha1")
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP Status: %d", res.StatusCode)
	}

	b, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	b = bytes.TrimSpace(b)

	b = b[:40]

	s := make([]byte, hex.DecodedLen(len(b)))

	if _, err := hex.Decode(s, b); err != nil {
		return nil, err
	}

	return s, nil
}

func shaOfFile(filename string) ([]byte, error) {
	r, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer r.Close()

	h := sha1.New()
	if _, err := io.Copy(h, r); err != nil {
		return nil, err
	}

	return h.Sum(nil), nil
}

type reader struct {
	sha []byte
	io.ReadCloser
	h hash.Hash
	u string
}

func (r *reader) Read(b []byte) (int, error) {
	n, err := r.ReadCloser.Read(b)
	if err == io.EOF {
		r.h.Write(b[:n])
		sha := r.h.Sum(nil)
		if !bytes.Equal(r.sha, sha) {
			return 0, fmt.Errorf("sha1 mismatch on %s (%s vs %s)",
				r.u,
				hex.EncodeToString(r.sha),
				hex.EncodeToString(sha))
		}
		return n, err
	} else if err != nil {
		return n, err
	}
	r.h.Write(b[:n])
	return n, err
}

func checkStatus(res *http.Response) error {
	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("http status: %d", res.StatusCode)
	}
	return nil
}

// Fetch ...
func Fetch(dst, url string) error {
	rs, err := getShaFor(url)
	if err != nil {
		return err
	}

	ls, err := shaOfFile(dst)
	if err == nil && bytes.Equal(ls, rs) {
		return nil
	}

	res, err := getWithCheck(url, rs)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	w, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer w.Close()

	if _, err := io.Copy(w, res.Body); err != nil {
		return err
	}

	return nil
}

func getWithCheck(url string, sum []byte) (*http.Response, error) {
	res, err := http.Get(url)
	if err != nil {
		return nil, err
	}

	if err := checkStatus(res); err != nil {
		return nil, err
	}

	res.Body = &reader{
		ReadCloser: res.Body,
		sha:        sum,
		h:          sha1.New(),
		u:          url,
	}

	return res, nil
}

// GetWithCheck ...
func GetWithCheck(url string) (*http.Response, error) {
	sum, err := getShaFor(url)
	if err != nil {
		return nil, err
	}

	return getWithCheck(url, sum)
}
