package cookiejar

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

// endOfTime is the time when session (non-persistent) cookies expire. This
// instant is representable in most date/time formats (not just Go's time.Time)
// and should be far enough in the future.
var endOfTime = time.Date(9999, 12, 31, 23, 59, 59, 0, time.UTC)

type Jar struct {
	cookies   map[string]map[string]*http.Cookie
	filename  string
	hash      string
}

type jarFile struct {
	Cookies map[string]map[string]*http.Cookie
	Hash    string
}

func Open(filename, hash string) (j *Jar, hashMatch bool, err error) {
	f := &jarFile{make(map[string]map[string]*http.Cookie), hash}

	if _, err := os.Stat(filename); err == nil {
		b, err := ioutil.ReadFile(filename)
		if err != nil {
			return nil, false, err
		}

		if err = json.Unmarshal(b, f); err != nil {
			return nil, false, err
		}

		hashMatch = hash == f.Hash
	}

	j = &Jar{
		cookies:  make(map[string]map[string]*http.Cookie),
		filename: filename,
		hash:     hash,
	}

	if hashMatch {
		j.cookies = f.Cookies
	}

	return j, hashMatch, nil
}

func (j *Jar) Flush() error {
	f := &jarFile{j.cookies, j.hash}

	b, err := json.Marshal(f)
	if err != nil {
		return err
	}

	if err = ioutil.WriteFile(j.filename, b, 0664); err != nil {
		return err
	}

	return nil
}

func (j *Jar) Cookies(u *url.URL) (cookies []*http.Cookie) {
	host := u.Host
	if i := strings.Index(host, ":"); i > 0 {
		host = host[:i]
	}

	if _, ok := j.cookies[host]; !ok {
		return cookies
	}

	now := time.Now()
	for _, c := range j.cookies[host] {
		if !c.Expires.IsZero() && !c.Expires.After(now) {
			delete(j.cookies[host], c.Name)
		} else {
			cookies = append(cookies, c)
		}
	}

	return cookies
}

func (j *Jar) SetCookies(u *url.URL, cookies []*http.Cookie) {
	host := u.Host
	if i := strings.Index(host, ":"); i > 0 {
		host = host[:i]
	}

	if _, ok := j.cookies[host]; !ok {
		j.cookies[host] = make(map[string]*http.Cookie)
	}

	// BUG(scjudd): This logic should keep track of whether a cookie should
	// persist or not; that is, if it is a session cookie. If so, it shouldn't
	// be written to the jarFile on Flush. This will probably require making a
	// new type, since there's no field to address persistence in http.Cookie.
	now := time.Now()
	for _, c := range cookies {
		switch {
		case c.MaxAge < 0:
			continue
		case c.MaxAge > 0:
			c.Expires = now.Add(time.Duration(c.MaxAge) * time.Second)
		case c.Expires.IsZero():
			c.Expires = endOfTime
		case !c.Expires.After(now):
			continue
		}
		j.cookies[host][c.Name] = c
	}
}
