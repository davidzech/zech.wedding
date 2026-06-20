package main

import (
	"bytes"
	"compress/gzip"
	_ "embed"
	"io"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strconv"
	"strings"
)

const target = "https://www.theknot.com/us/anne-nguyen-and-david-zech-2027-06-05"

//go:embed patch.html
var patch string

func main() {
	targetURL, err := url.Parse(target)
	if err != nil {
		log.Fatal(err)
	}

	proxy := &httputil.ReverseProxy{}
	proxy.Rewrite = func(r *httputil.ProxyRequest) {
		r.SetURL(targetURL)
		r.Out.Host = targetURL.Host
		r.Out.URL.Path = targetURL.Path
		r.Out.URL.RawQuery = targetURL.RawQuery
		// Tell the server we can handle gzip so we get compressed responses,
		// but we'll decompress them ourselves before injecting the script.
		r.Out.Header.Set("Accept-Encoding", "gzip")
	}

	proxy.ModifyResponse = func(resp *http.Response) error {
		if !strings.Contains(resp.Header.Get("Content-Type"), "text/html") {
			return nil
		}

		// Strip headers that would block our injected inline script.
		resp.Header.Del("Content-Security-Policy")
		resp.Header.Del("X-Frame-Options")

		// Decompress if needed.
		var reader io.Reader = resp.Body
		if resp.Header.Get("Content-Encoding") == "gzip" {
			gz, err := gzip.NewReader(resp.Body)
			if err != nil {
				return err
			}
			defer gz.Close()
			reader = gz
			resp.Header.Del("Content-Encoding")
		}

		body, err := io.ReadAll(reader)
		resp.Body.Close()
		if err != nil {
			return err
		}

		// Neutralize the OneTrust cookie-consent SDK. An inline snippet in the
		// page (id="Union__consent-management__snippet") creates a
		// <script src=".../otSDKStub.js"> at runtime, which builds the cookie
		// popup. Pointing that URL at an empty data URI means the SDK never
		// loads (no popup) while window.UnionConsentManagement stays defined,
		// so the other inline scripts that call it don't throw.
		body = bytes.ReplaceAll(body,
			[]byte("https://cdn.cookielaw.org/scripttemplates/otSDKStub.js"),
			[]byte("data:text/javascript,"))

		// Inject our script before </body>.
		modified := bytes.Replace(body, []byte("</body>"), []byte(patch+"</body>"), 1)

		resp.Body = io.NopCloser(bytes.NewReader(modified))
		resp.ContentLength = int64(len(modified))
		resp.Header.Set("Content-Length", strconv.Itoa(len(modified)))
		return nil
	}

	addr := ":" + port()
	log.Printf("Proxying %s -> %s", addr, target)
	log.Fatal(http.ListenAndServe(addr, proxy))
}

func port() string {
	if p := os.Getenv("PORT"); p != "" {
		return p
	}
	return "8080"
}
