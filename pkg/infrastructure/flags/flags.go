package flags

import (
	"bytes"
	"fmt"
	"net/url"
	"os"
	"strings"
)

type Uri struct {
	Key string
	Val url.URL
}

func (u *Uri) String() string {
	val := u.Val.String()
	if val == "" {
		return fmt.Sprintf("%s", u.Key)
	}
	return fmt.Sprintf("%s:%s", u.Key, val)
}

func (u *Uri) Set(value string) error {
	s := strings.SplitN(value, ":", 2)
	if s[0] == "" {
		return fmt.Errorf("missing uri key in '%s'", value)
	}
	u.Key = s[0]
	if len(s) > 1 && s[1] != "" {
		e := os.ExpandEnv(s[1])
		uri, err := url.Parse(e)
		if err != nil {
			return err
		}
		u.Val = *uri
	}
	return nil
}

type Uris []Uri

func (us *Uris) String() string {
	var b bytes.Buffer
	b.WriteString("[")
	for i, u := range *us {
		if i > 0 {
			b.WriteString(" ")
		}
		b.WriteString(u.String())
	}
	b.WriteString("]")
	return b.String()
}

func (us *Uris) Set(value string) error {
	var u Uri
	if err := u.Set(value); err != nil {
		return err
	}
	*us = append(*us, u)
	return nil
}

func (us *Uris) Type() string {
	return fmt.Sprintf("%T", us)
}
