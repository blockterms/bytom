package addresscallbacks

import (
	"encoding/json"
	"os"
	"testing"
	"time"

	dbm "github.com/tendermint/tmlibs/db"
)

func TestGetSetTxDBValue(t *testing.T) {
	testDB := dbm.NewDB("txdb", "leveldb", "temp")
	defer os.RemoveAll("temp")
	timeNow := time.Now().Unix()
	sampleValue := &TxDBValue{Unixtime: timeNow}
	valueIn, err := json.Marshal(sampleValue)
	if err != nil {
		t.Errorf("There should be no error marshalling a sample txDBValue struct")
	}
	k := []byte("txid")
	testDB.Set(k, valueIn)
	valueOut := TxDBValue{}
	sampleValueOut := testDB.Get(k)
	errDecode := json.Unmarshal(sampleValueOut, &valueOut)
	if errDecode != nil {
		t.Errorf("There should be no error unmarshalling a sample txDBValue struct")
	}

	if timeNow != valueOut.Unixtime {
		t.Errorf("value Set in DB is not the same as Get value")
	}

}
