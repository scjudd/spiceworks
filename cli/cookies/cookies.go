package cookies

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
)

type Jar struct {
	cookies  map[string]map[string]string
	filename string
}

func Open(name string) (*Jar, error) {
	jar := &Jar{
		cookies:  make(map[string]map[string]string),
		filename: name,
	}

	if _, err := os.Stat(name); err != nil {
		file, err := os.Create(name)
		if err != nil {
			return nil, err
		}
		defer file.Close()
		_, err = file.WriteString("{}")
		if err != nil {
			return nil, err
		}
	} else {
		b, err := ioutil.ReadFile(name)
		if err != nil {
			return nil, err
		}
		if err = json.Unmarshal(b, &jar.cookies); err != nil {
			return nil, err
		}
	}

	return jar, nil
}

func (j *Jar) Flush() error {
	file, err := os.OpenFile(j.filename, os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		return err
	}

	b, err := json.Marshal(j.cookies)
	if err != nil {
		return err
	}

	_, err = file.Write(b)
	if err != nil {
		return err
	}

	return nil
}

func (j *Jar) Cookies(u *url.URL) (cookies []*http.Cookie) {
	if submap, ok := j.cookies[u.Host]; ok {
		for name, value := range submap {
			cookies = append(cookies, &http.Cookie{Name: name, Value: value})
		}
	}
	return cookies
}

func (j *Jar) SetCookies(u *url.URL, cookies []*http.Cookie) {
	if _, ok := j.cookies[u.Host]; !ok {
		j.cookies[u.Host] = make(map[string]string)
	}
	for _, cookie := range cookies {
		j.cookies[u.Host][cookie.Name] = cookie.Value
	}
}
