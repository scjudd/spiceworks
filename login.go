package spiceworks

import (
	"bytes"
	"errors"
	"io/ioutil"
	"launchpad.net/xmlpath"
	"net/url"
)

// Login performs the necessary HTTP requests in order to populate the Client's
// underlying http.Client's CookieJar. In other words, this method only needs
// to be called if the CookieJar doesn't already contain valid cookie
// information.
func (c *Client) Login() error {
	loginUrl := c.BaseUrl + "login"

	// First, we must grab an authenticity_token by requesting the login page and
	// parsing it out of the response.

	resp, err := c.HttpClient.Get(loginUrl)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	xmlroot, err := xmlpath.ParseHTML(resp.Body)
	if err != nil {
		return err
	}

	path, err := xmlpath.Compile("//input[@name='authenticity_token']/@value")
	if err != nil {
		return err
	}

	authenticity_token, ok := path.String(xmlroot)
	if !ok {
		return errors.New("Couldn't find authenticity_token.")
	}

	// Now that we've got an authenticity_token, we continue with the actual
	// login process.

	data := url.Values{}
	data.Set("authenticity_token", authenticity_token)
	data.Set("user[email]", c.Email)
	data.Set("user[password]", c.Password)
	data.Set("user[remember]", "1")

	resp, err = c.HttpClient.PostForm(loginUrl, data)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if bytes.Contains(body, []byte("Login failed")) {
		return errors.New("Login failed!")
	}

	return nil
}
