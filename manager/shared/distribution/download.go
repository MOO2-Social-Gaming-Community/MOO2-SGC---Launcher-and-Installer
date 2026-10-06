package distribution

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"syscall"
	"time"
)

type Failure struct {
	Kind     string `json:"kind"`
	Provider string `json:"provider,omitempty"`
	Message  string `json:"message"`
}

func (f Failure) Error() string { return f.Kind + ": " + f.Message }
func safeFailureMessage(e error) string {
	var u *url.Error
	if errors.As(e, &u) {
		return "HTTPS request failed (" + classify(u.Err) + ")"
	}
	return e.Error()
}
func classify(e error) string {
	var ne net.Error
	var de *net.DNSError
	switch {
	case errors.Is(e, io.ErrUnexpectedEOF):
		return "incomplete"
	case errors.Is(e, syscall.ENOSPC):
		return "disk-full"
	case errors.Is(e, os.ErrPermission):
		return "permission"
	case errors.As(e, &de):
		return "dns"
	case errors.As(e, &ne) && ne.Timeout():
		return "timeout"
	case errors.Is(e, context.Canceled):
		return "cancelled"
	}
	return "transfer"
}

type Client struct {
	Trust Trust
	HTTP  *http.Client
	Log   func(Failure)
	// Local binary policy, never supplied by a downloaded manifest.
	MinimumLauncherVersion string
	LauncherPlatform       string
	// Tests can inject a loopback HTTP transport; no CLI or manifest can enable this.
	fixture bool
}

func NewClient(t Trust) *Client {
	c := &Client{Trust: t}
	c.HTTP = &http.Client{Timeout: 5 * time.Minute, CheckRedirect: func(r *http.Request, via []*http.Request) error {
		if len(via) > 8 {
			return errors.New("redirect limit")
		}
		return t.CheckURL(r.URL.String())
	}}
	return c
}
func (c *Client) checkURL(u string) error {
	if c.fixture {
		return nil
	}
	return c.Trust.CheckURL(u)
}
func (c *Client) emit(f Failure) {
	if c.Log != nil {
		c.Log(f)
	}
}
func (c *Client) get(ctx context.Context, u string, limit int64) ([]byte, error) {
	if e := c.checkURL(u); e != nil {
		return nil, e
	}
	req, e := http.NewRequestWithContext(ctx, "GET", u, nil)
	if e != nil {
		return nil, e
	}
	req.Header.Set("Accept-Encoding", "identity")
	req.Header.Set("Cache-Control", "no-cache")
	r, e := c.HTTP.Do(req)
	if e != nil {
		return nil, e
	}
	defer r.Body.Close()
	if r.StatusCode != http.StatusOK {
		return nil, Failure{"http", "", fmt.Sprintf("HTTP %d", r.StatusCode)}
	}
	if r.ContentLength > limit {
		return nil, errors.New("response size limit")
	}
	b, e := io.ReadAll(io.LimitReader(r.Body, limit+1))
	if e != nil {
		return nil, e
	}
	if int64(len(b)) > limit {
		return nil, errors.New("response size limit")
	}
	return b, nil
}

// FetchManifest only accepts a signed, fresh release. Fallback does not lower
// trusted revision state. Signature failures are logged even if a mirror works.
func (c *Client) FetchManifest(ctx context.Context, prior TrustedState, now time.Time) (VerifiedManifest, error) {
	if e := c.Trust.Validate(); e != nil {
		return VerifiedManifest{}, e
	}
	ep := append([]Endpoint{}, c.Trust.Endpoints...)
	sort.SliceStable(ep, func(i, j int) bool { return ep[i].Priority < ep[j].Priority })
	if len(ep) == 0 {
		return VerifiedManifest{}, errors.New("no production feed configured; use the signed offline test kit or provision release endpoints")
	}
	var failures []string
	for _, m := range ep {
		b, e := c.get(ctx, m.URL, MaxMetadata)
		var s []byte
		if e == nil {
			s, e = c.get(ctx, m.SignatureURL, 4096)
		}
		if e == nil {
			var v VerifiedManifest
			v, e = c.Trust.VerifyManifest(b, s, now, prior, true)
			if e == nil {
				e = v.RequireLauncher(c.LauncherPlatform, c.MinimumLauncherVersion)
			}
			if e == nil {
				return v, nil
			}
		}
		f := Failure{classify(e), m.Provider, safeFailureMessage(e)}
		c.emit(f)
		failures = append(failures, m.Provider+": "+safeFailureMessage(e))
		if ctx.Err() != nil {
			break
		}
	}
	return VerifiedManifest{}, fmt.Errorf("no trusted release metadata available: %s", strings.Join(failures, "; "))
}
func rangeValid(h string, start, total int64) bool {
	// Require exactly the remaining object, not an arbitrary partial chunk.
	var a, b, n int64
	_, e := fmt.Sscanf(h, "bytes %d-%d/%d", &a, &b, &n)
	return e == nil && h == fmt.Sprintf("bytes %d-%d/%d", a, b, n) && a == start && b == total-1 && n == total
}
func (c *Client) transfer(ctx context.Context, u, partial string, p Package) error {
	if e := c.checkURL(u); e != nil {
		return e
	}
	if e := NoLinks(partial); e != nil {
		return e
	}
	var offset int64
	if st, e := os.Stat(partial); e == nil {
		if !st.Mode().IsRegular() {
			return errors.New("partial is not a file")
		}
		offset = st.Size()
	} else if !os.IsNotExist(e) {
		return e
	}
	if offset > p.Size {
		if e := os.Remove(partial); e != nil {
			return e
		}
		offset = 0
	}
	if offset == p.Size {
		return VerifyFile(partial, p)
	}
	req, e := http.NewRequestWithContext(ctx, "GET", u, nil)
	if e != nil {
		return e
	}
	req.Header.Set("Accept-Encoding", "identity")
	if offset > 0 {
		req.Header.Set("Range", "bytes="+strconv.FormatInt(offset, 10)+"-")
	}
	resp, e := c.HTTP.Do(req)
	if e != nil {
		return e
	}
	defer resp.Body.Close()
	if resp.Header.Get("Content-Encoding") != "" && resp.Header.Get("Content-Encoding") != "identity" {
		return errors.New("encoded package transfer refused")
	}
	switch resp.StatusCode {
	case http.StatusOK:
		offset = 0 // Range unsupported: restart safely, never append a whole object.
	case http.StatusPartialContent:
		if !rangeValid(resp.Header.Get("Content-Range"), offset, p.Size) {
			return Failure{"range", "", "invalid Content-Range"}
		}
	case http.StatusRequestedRangeNotSatisfiable:
		_ = os.Remove(partial)
		return Failure{"range", "", "server rejected range; partial reset for retry"}
	default:
		return Failure{"http", "", fmt.Sprintf("HTTP %d", resp.StatusCode)}
	}
	expected := p.Size - offset
	if resp.ContentLength >= 0 && resp.ContentLength != expected {
		return Failure{"size", "", "server byte count differs from signed metadata"}
	}
	flags := os.O_CREATE | os.O_WRONLY
	if offset == 0 {
		flags |= os.O_TRUNC
	} else {
		flags |= os.O_APPEND
	}
	out, e := os.OpenFile(partial, flags, 0600)
	if e != nil {
		return e
	}
	n, e := io.Copy(out, io.LimitReader(resp.Body, expected+1))
	se := out.Sync()
	ce := out.Close()
	if n > expected {
		_ = os.Remove(partial)
		return Failure{"size", "", "download exceeded signed byte count"}
	}
	if e != nil {
		return e
	}
	if se != nil {
		return se
	}
	if ce != nil {
		return ce
	}
	if n != expected {
		return Failure{"incomplete", "", "interrupted transfer; partial retained for resume"}
	}
	if e = VerifyFile(partial, p); e != nil {
		_ = os.Remove(partial)
		return Failure{"integrity", "", e.Error()}
	}
	return nil
}

// Download uses a content-addressed cache. Only authenticated, exact bytes get
// the final filename. A poisoned partial is discarded before mirror fallback.
func (c *Client) Download(ctx context.Context, p Package, cache string) (string, error) {
	if e := c.Trust.VerifyPackage(p); e != nil {
		return "", e
	}
	if e := NoLinks(cache); e != nil {
		return "", e
	}
	if e := os.MkdirAll(cache, 0700); e != nil {
		return "", e
	}
	dest := filepath.Join(cache, p.SHA256+".zip")
	if e := VerifyFile(dest, p); e == nil {
		return dest, nil
	}
	mirrors := append([]Mirror{}, p.Mirrors...)
	sort.SliceStable(mirrors, func(i, j int) bool { return mirrors[i].Priority < mirrors[j].Priority })
	if len(mirrors) == 0 {
		return "", errors.New("package has no network mirrors; signed offline import is required")
	}
	partial := filepath.Join(cache, p.SHA256+".part")
	var failures []string
	for _, m := range mirrors {
		for attempt := 0; attempt < 2; attempt++ {
			e := c.transfer(ctx, m.URL, partial, p)
			if e == nil {
				if e = c.Trust.VerifyPackage(p); e != nil {
					return "", e
				}
				if e = replace(partial, dest); e != nil {
					return "", e
				}
				return dest, nil
			}
			kind := classify(e)
			var f Failure
			if errors.As(e, &f) {
				kind = f.Kind
			}
			c.emit(Failure{kind, m.Provider, safeFailureMessage(e)})
			failures = append(failures, m.Provider+": "+safeFailureMessage(e))
			if kind == "disk-full" || kind == "permission" || ctx.Err() != nil {
				return "", e
			}
			// Never keep an entire bad object for the next attempt.
			if st, se := os.Stat(partial); se == nil && st.Size() >= p.Size {
				_ = os.Remove(partial)
			}
			if kind == "integrity" || kind == "size" {
				_ = os.Remove(partial)
				break
			}
		}
	}
	return "", fmt.Errorf("all package mirrors failed: %s", strings.Join(failures, "; "))
}
