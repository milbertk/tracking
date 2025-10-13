package tracking

import (
	"encoding/json"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/oschwald/geoip2-golang"
)

type Collector struct {
	geo *geoip2.Reader // nil-safe: if nil, country will be "Unknown" unless a CDN header is present
}

func NewCollector(geoLite2Path string) (*Collector, error) {
	db, err := geoip2.Open(geoLite2Path) // e.g., ./GeoLite2-Country.mmdb
	if err != nil {
		return nil, err
	}
	return &Collector{geo: db}, nil
}

func (c *Collector) Close() error {
	if c.geo != nil {
		return c.geo.Close()
	}
	return nil
}

type Info struct {
	IP          string `json:"ip"`
	Platform    string `json:"platform"`     // OS guess (best-effort from User-Agent)
	Browser     string `json:"browser"`      // Browser guess (best-effort from User-Agent)
	CountryCode string `json:"country_code"` // GeoIP or CDN header
	GMTTime     string `json:"gmt_time"`     // client UTC offset (e.g., "-360" minutes) — sent by client
	Lang        string `json:"lang"`         // first of Accept-Language, e.g., "es-CR"
	UserAgent   string `json:"user_agent"`   // full UA for debugging
	RequestTime string `json:"request_time"` // server time (for reference)
}

// Extract gathers client metadata from *http.Request.
// Country resolution order:
//  1. If CDN sets CF-IPCountry, use that.
//  2. Else if GeoIP DB loaded, map IP -> ISO country code.
//  3. Else "Unknown".
func (c *Collector) Extract(r *http.Request) Info {
	ua := r.Header.Get("User-Agent")
	browser, platform := parseUA(ua)
	ip := clientIP(r)
	lang := firstLang(r.Header.Get("Accept-Language"))

	// 1) Prefer CDN header if present (e.g., Cloudflare)
	country := strings.TrimSpace(r.Header.Get("CF-IPCountry"))

	// 2) GeoIP fallback
	if country == "" && c.geo != nil && ip != "" {
		if p := net.ParseIP(ip); p != nil {
			if rec, err := c.geo.Country(p); err == nil && rec != nil && rec.Country.IsoCode != "" {
				country = rec.Country.IsoCode
			}
		}
	}
	if country == "" {
		country = "Unknown"
	}

	// Client GMT offset (UTC minutes), the browser should send this:
	//   X-Client-UTC-Offset: String(-new Date().getTimezoneOffset())
	gmt := strings.TrimSpace(r.Header.Get("X-Client-UTC-Offset"))

	return Info{
		IP:          ip,
		Platform:    platform,
		Browser:     browser,
		CountryCode: country,
		GMTTime:     gmt,
		Lang:        lang,
		UserAgent:   ua,
		RequestTime: time.Now().Format("2006-01-02 15:04:05"),
	}
}

// JSON pretty-prints Info.
func (i Info) JSON() string {
	b, _ := json.MarshalIndent(i, "", "  ")
	return string(b)
}

// ----------------- helpers -----------------

func clientIP(r *http.Request) string {
	// X-Forwarded-For may have multiple IPs: client, proxy1, proxy2...
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		for _, p := range strings.Split(xff, ",") {
			p = strings.TrimSpace(p)
			if net.ParseIP(p) != nil {
				return p
			}
		}
	}
	// Fallback to RemoteAddr (host:port)
	host, _, err := net.SplitHostPort(strings.TrimSpace(r.RemoteAddr))
	if err == nil && net.ParseIP(host) != nil {
		return host
	}
	if ip := net.ParseIP(r.RemoteAddr); ip != nil {
		return ip.String()
	}
	return ""
}

func parseUA(ua string) (browser, platform string) {
	u := strings.ToLower(ua)
	// browser
	switch {
	case strings.Contains(u, "edg/"):
		browser = "Edge"
	case strings.Contains(u, "chrome/"):
		browser = "Chrome"
	case strings.Contains(u, "firefox/"):
		browser = "Firefox"
	case strings.Contains(u, "safari/"):
		browser = "Safari"
	default:
		browser = "Unknown"
	}
	// platform
	switch {
	case strings.Contains(u, "windows"):
		platform = "Windows"
	case strings.Contains(u, "macintosh") || strings.Contains(u, "mac os"):
		platform = "macOS"
	case strings.Contains(u, "android"):
		platform = "Android"
	case strings.Contains(u, "iphone") || strings.Contains(u, "ipad") || strings.Contains(u, "ios"):
		platform = "iOS"
	case strings.Contains(u, "linux"):
		platform = "Linux"
	default:
		platform = "Unknown"
	}
	return
}

func firstLang(al string) string {
	al = strings.TrimSpace(al)
	if al == "" {
		return ""
	}
	if i := strings.IndexByte(al, ','); i > 0 {
		return al[:i]
	}
	return al
}

/*

package main

import (
	"fmt"
	"log"
	"net/http"

	"your/module/clientmeta"
)

func main() {
	// Download MaxMind GeoLite2 Country DB and point to it here:
	// https://dev.maxmind.com/geoip/geolite2-free-geolocation-data
	collector, err := clientmeta.NewCollector("./GeoLite2-Country.mmdb")
	if err != nil {
		log.Fatal("GeoIP load error:", err)
	}
	defer collector.Close()

	http.HandleFunc("/whoami", func(w http.ResponseWriter, r *http.Request) {
		// If your frontend sends client offset, allow it in CORS preflight:
		// Access-Control-Allow-Headers: Authorization, Content-Type, X-Client-UTC-Offset

		info := collector.Extract(r)

		// Return JSON
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(info.JSON()))

		// Also log to server
		fmt.Println(info.JSON())
	})

	log.Println("listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}


*/
