package main

import (
	"code.google.com/p/gorilla/securecookie"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
)

var (
	config = struct {
		GoogleOauthClientId     string
		GoogleOauthClientSecret string
		CookieAuthKeyHexStr     string
		CookieEncrKeyHexStr     string
		AdminEmail              string
		AdminId                 string
	}{
		"", "",
		"", "",
		"", "",
	}

	cookieAuthKey []byte
	cookieEncrKey []byte
	secureCookie  *securecookie.SecureCookie
	cookieName    = "buq2_cookie"
)

func getGoogleOauthClientIdAndSecret() (string, string) {
	return config.GoogleOauthClientId, config.GoogleOauthClientSecret
}

func readAuthConfig() error {
	b, err := ioutil.ReadFile("auth_config.json")
	if err != nil {
		return err
	}
	err = json.Unmarshal(b, &config)
	if err != nil {
		return err
	}
	cookieAuthKey, err = hex.DecodeString(config.CookieAuthKeyHexStr)
	if err != nil {
		return err
	}
	cookieEncrKey, err = hex.DecodeString(config.CookieEncrKeyHexStr)
	if err != nil {
		return err
	}
	secureCookie = securecookie.New(cookieAuthKey, cookieEncrKey)
	// verify auth/encr keys are correct
	val := map[string]string{
		"foo": "bar",
	}
	_, err = secureCookie.Encode(cookieName, val)
	if err != nil {
		// for convenience, if the auth/encr keys are not set,
		// generate valid, random value for them
		auth := securecookie.GenerateRandomKey(32)
		encr := securecookie.GenerateRandomKey(32)
		fmt.Printf("cookieAuthKey: %s\ncookieEncrKey: %s\n", hex.EncodeToString(auth), hex.EncodeToString(encr))
	}

	return err
}
