package addresscallbacks

import (
	"encoding/json"
	"net/url"
	"regexp"
	"strings"
	"unicode/utf8"

	"github.com/bytom/errors"
	dbm "github.com/tendermint/tmlibs/db"
)

const maxURLRuneCount = 2083
const minURLRuneCount = 3

// Basic regular expression for validating urls
const (
	URL string = `^(?:http(s)?:\/\/)?[\w.-]+(?:\.[\w\.-]+)+[\w\-\._~:/?#[\]@!\$&'\(\)\*\+,;=.]+$`
)

var (
	rxURL = regexp.MustCompile(URL)
	// ErrBadAddress is returned when Add is called on an invalid address string.
	ErrBadAddress = errors.New("invalid address")
	// ErrNotValidURL is returned when a callback url is not valid.
	ErrNotValidURL = errors.New("not a valid url")
	// ErrDuplicateURL is returned when Add is called with a duplicate url.
	ErrDuplicateURL = errors.New(" url is already in the list")
	// ErrNoCallbacks is returned when Delete is called on address with no callbacks
	ErrNoCallbacks = errors.New("No callbacks listed for address")
	// ErrCallbackNotFound is returned when Delete is called on address with no callbacks
	ErrCallbackNotFound = errors.New("Cannot find the callback url for deletion")
)

// CallbackStore stores callbacks for watching address txs.
type CallbackStore struct {
	DB dbm.DB
}

// NewStore creates and returns a new Store object.
func NewStore(db dbm.DB) *CallbackStore {
	return &CallbackStore{
		DB: db,
	}
}

// Add adds callback url for a given address.
func (cs *CallbackStore) Add(address string, callbackURL string) (bool, error) {
	if len(address) < 42 {
		return false, ErrBadAddress
	}
	if !IsURL(callbackURL) {
		return false, ErrNotValidURL
	}
	key := []byte(address)
	cbs := make([]string, 0)
	callbacksJSON := cs.DB.Get(key)
	if callbacksJSON != nil {
		err := json.Unmarshal(callbacksJSON, &cbs)
		if err != nil {
			return false, err
		}
		anyDups := map[string]bool{}
		for _, element := range cbs {
			anyDups[element] = true
		}
		if anyDups[callbackURL] {
			return false, ErrDuplicateURL
		}
	}
	cbs = append(cbs, callbackURL)
	value, err := json.Marshal(cbs)
	if err != nil {
		return false, err
	}
	cs.DB.Set(key, value)

	return true, nil
}

// List lists all callback urls for given address.
func (cs *CallbackStore) List(address string) ([]string, error) {
	if len(address) < 42 {
		return nil, ErrBadAddress
	}
	key := []byte(address)
	callbacksJSON := cs.DB.Get(key)
	if callbacksJSON != nil {
		cbs := make([]string, 0)
		err := json.Unmarshal(callbacksJSON, &cbs)
		if err != nil {
			return nil, err
		}
		return cbs, nil
	}
	return []string{}, nil
}

// Delete deletes a given callback from the callbacks list for a given address
func (cs *CallbackStore) Delete(address, callback string) error {
	if len(address) < 42 {
		return ErrBadAddress
	}
	key := []byte(address)
	callbacksJSON := cs.DB.Get(key)
	if callbacksJSON == nil {
		return ErrNoCallbacks
	}
	cbs := make([]string, 0)
	err := json.Unmarshal(callbacksJSON, &cbs)
	if err != nil {
		return err
	}
	newCallbacks := []string{}
	for _, el := range cbs {
		if el != callback {
			newCallbacks = append(newCallbacks, el)
		}
	}
	if len(cbs) == len(newCallbacks) {
		return ErrCallbackNotFound
	}
	if len(newCallbacks) == 0 {
		cs.DB.Delete(key)
	} else {
		value, merr := json.Marshal(newCallbacks)
		if merr != nil {
			return merr
		}
		cs.DB.Set(key, value)
	}

	return nil
}

// IsURL check if the string is an URL.
// extracted from package https://github.com/asaskevich/govalidator
func IsURL(str string) bool {
	if str == "" || utf8.RuneCountInString(str) >= maxURLRuneCount || len(str) <= minURLRuneCount || strings.HasPrefix(str, ".") {
		return false
	}
	strTemp := str
	if strings.Contains(str, ":") && !strings.Contains(str, "://") {
		// support no indicated urlscheme but with colon for port number
		// http:// is appended so url.Parse will succeed, strTemp used so it does not impact rxURL.MatchString
		strTemp = "http://" + str
	}
	u, err := url.Parse(strTemp)
	if err != nil {
		return false
	}
	if strings.HasPrefix(u.Host, ".") {
		return false
	}
	if u.Host == "" && (u.Path != "" && !strings.Contains(u.Path, ".")) {
		return false
	}
	return rxURL.MatchString(str)
}
