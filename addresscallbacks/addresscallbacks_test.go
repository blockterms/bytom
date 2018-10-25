package addresscallbacks

import (
	"os"
	"testing"

	"github.com/bytom/errors"
	dbm "github.com/tendermint/tmlibs/db"
)

func TestIsURL(t *testing.T) {
	if !IsURL("https://bytom.com") {
		t.Errorf("%s is a valid url", "https://bytom.com")
	}
	if IsURL("https//bytom.com") {
		t.Errorf("%s is not a valid url", "https//bytom.com")
	}
}
func TestAdd(t *testing.T) {
	testDB := dbm.NewDB("addressdb", "leveldb", "temp")
	defer os.RemoveAll("temp")
	cs := NewStore(testDB)

	cases := []struct {
		address, url string
		want         error
	}{
		{"sample address should be of length more than 42", "https://bytom.com", nil},
		{"sample address should be of length more than 42", "https://bytom.com", ErrDuplicateURL},
		{"", "https://blahblah.com", ErrBadAddress},
		{"sample address should be of length more than 42", "bad url", ErrNotValidURL},
	}

	for _, c := range cases {
		_, err := cs.Add(c.address, c.url)
		if errors.Root(err) != c.want {
			t.Errorf("Add(%s, %s) error = %s want %s", c.address, c.url, err, c.want)
		}
	}
}

func TestList(t *testing.T) {
	testDB := dbm.NewDB("addressdb", "leveldb", "temp")
	defer os.RemoveAll("temp")
	cs := NewStore(testDB)
	sampleAddress := "sample address should be of length more than 42 - 1"
	cbInp := []string{
		"https://bytom.com/1",
		"https://bytom.com/2",
		"https://bytom.com/3"}
	cs.Add(sampleAddress, cbInp[0])
	cs.Add(sampleAddress, cbInp[1])
	cs.Add(sampleAddress, cbInp[2])

	callbacks, err := cs.List(sampleAddress)
	if err != nil {
		t.Errorf("There should be no error getting list of callbacks")
	}
	if len(callbacks) != 3 {
		t.Errorf("Expecting 3 callback urls for sample address")
	}
	if callbacks[0] != cbInp[0] ||
		callbacks[1] != cbInp[1] ||
		callbacks[2] != cbInp[2] {
		t.Errorf("List output is not the same as input")
	}
}

func TestDelete(t *testing.T) {
	testDB := dbm.NewDB("addressdb", "leveldb", "temp")
	defer os.RemoveAll("temp")
	cs := NewStore(testDB)
	sampleAddress := "sample address should be of length more than 42 - 1"
	cbInp := []string{
		"https://bytom.com/1",
		"https://bytom.com/2",
		"https://bytom.com/3"}
	cs.Add(sampleAddress, cbInp[0])
	cs.Add(sampleAddress, cbInp[1])
	cs.Add(sampleAddress, cbInp[2])
	cases := []struct {
		address, url string
		want         error
	}{
		{"blahblah", "blahblah", ErrBadAddress},
		{sampleAddress, "https://bytom.com", ErrCallbackNotFound},
		{sampleAddress, "https://bytom.com/2", nil},
		{sampleAddress, "https://bytom.com/3", nil},
		{sampleAddress, "https://bytom.com/1", nil},
	}

	for _, c := range cases {
		err := cs.Delete(c.address, c.url)
		if errors.Root(err) != c.want {
			t.Errorf("Delete(%s, %s) error = %s want %s", c.address, c.url, err, c.want)
		}
	}

}
