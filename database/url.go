package database

import (
	"errors"
	"fmt"
	"net/url"
)

type URL struct {
	*url.URL
}

func URLMustParse(ur string) URL {
	l, err := url.Parse(ur)
	if err != nil {
		panic(err)
	}
	return URL{l}
}

func (u URL) String() string {
	if u.URL != nil {
		return u.URL.String()
	}
	return ""
}

func (u *URL) Scan(value any) error {
	if u == nil {
		return errors.New("destination pointer is nil")
	}
	var err error = nil
	switch v := value.(type) {
	case string:
		(*u).URL, err = url.Parse(v)
	case []byte:
		(*u).URL, err = url.Parse(string(v))
	case nil:
		(*u).URL = nil
	default:
		return fmt.Errorf("unsupported Scan, storing driver.Value type %T into type %T", value, u)
	}
	return err
}
