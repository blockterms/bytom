package addresscallbacks

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"net/http"
	"time"

	"github.com/bytom/blockchain/query"
	cfg "github.com/bytom/config"
	"github.com/bytom/protocol/bc/types"
	"github.com/bytom/wallet"
	log "github.com/sirupsen/logrus"
	dbm "github.com/tendermint/tmlibs/db"
)

const (
	maxTxChanSize = 10000
	maxTTL        = 48 * 60 * 60 // 48 hours
)

// TxListener listens new txs from peers and notifies the
// interested listeners.
type TxListener struct {
	CBStore CallbackStore
	txDB    dbm.DB
	TxCh    chan *types.Tx
}

// NewTxListener creates and returns a new TxListener object.
func NewTxListener(config *cfg.Config, cbStore *CallbackStore) *TxListener {
	txDB := dbm.NewDB("txlistener", config.DBBackend, config.DBDir())
	txListener := &TxListener{
		CBStore: *cbStore,
		txDB:    txDB,
		TxCh:    make(chan *types.Tx, maxTxChanSize),
	}
	go txListener.TxListenerLoop()
	go txListener.CleanUpOldTx(maxTTL)
	return txListener
}

// TxListenerLoop is constantly listening for new txs
func (txl *TxListener) TxListenerLoop() {
	for {
		select {
		case newTx := <-txl.TxCh:
			txl.ProcessTx(newTx)
		}
	}
}

// TxDBValue is a datatype for storing a value to txid as key
type TxDBValue struct {
	Unixtime int64
}

// CleanUpOldTx A function that cleans up the leveldb transactions
// that are older than 48 hours.
func (txl *TxListener) CleanUpOldTx(maxTTL int64) {
	for now := range time.Tick(time.Hour) {
		log.Info("Cleaning up txlistener Database")
		iter := txl.txDB.Iterator()
		for iter.Next() {
			key := iter.Key()
			value := iter.Value()
			valueOut := TxDBValue{}
			err := json.Unmarshal(value, &valueOut)
			if err != nil && now.Unix()-valueOut.Unixtime > int64(maxTTL) {
				txl.txDB.Delete(key)
			}
		}
	}
}

// UpdateTxDB with txID
func (txl *TxListener) UpdateTxDB(tIDBytes []byte) {
	timeNow := time.Now().Unix()
	sampleValue := &TxDBValue{Unixtime: timeNow}
	valueIn, _ := json.Marshal(sampleValue)
	txl.txDB.Set(tIDBytes, valueIn)
}

// CallbackData sent to the callback listener urls in http POST
type CallbackData struct {
	AssetID string `json:"asset_id"`
	Amount  uint64 `json:"amount"`
	Address string `json:"address"`
	TxID    string `json:"tx_id"`
}

// ProcessTx Will handle the tx
func (txl *TxListener) ProcessTx(newTx *types.Tx) {
	log.WithField("Txid", newTx.ID.String()).Info("Processing a new Tx")
	tIDBytes := []byte(newTx.ID.String())
	txInDB := txl.txDB.Get(tIDBytes)
	if txInDB == nil {
		txl.UpdateTxDB(tIDBytes)
		allInputs, allOutputs := constructInputsOutputs(newTx)
		for _, outp := range allOutputs {
			urls, err := txl.CBStore.List(outp.Address)
			if err == nil && len(urls) > 0 {
				alsoInput := isOutputAddressInInputs(allInputs, outp.Address)
				if !alsoInput {
					callbackData := &CallbackData{
						AssetID: outp.AssetID.String(),
						Amount:  outp.Amount,
						Address: outp.Address,
						TxID:    newTx.ID.String(),
					}
					// we found an address list to notify
					for _, callbackAddress := range urls {
						//Call all of the end points concurrently
						go Callback(callbackAddress, callbackData)
					}
				}
			}
		}
	}
}

func isOutputAddressInInputs(inputs []*query.AnnotatedInput, address string) bool {
	for _, inp := range inputs {
		if inp.Address == address {
			return true
		}
	}
	return false
}

func constructInputsOutputs(newTx *types.Tx) ([]*query.AnnotatedInput, []*query.AnnotatedOutput) {
	// The methods to convert txs are in wallet.
	emptyWallet := &wallet.Wallet{}
	allInputs := make([]*query.AnnotatedInput, 0, len(newTx.Inputs))
	allOutputs := make([]*query.AnnotatedOutput, 0, len(newTx.Outputs))

	for i := range newTx.Inputs {
		allInputs = append(allInputs, emptyWallet.BuildAnnotatedInput(newTx, uint32(i)))
	}
	for i := range newTx.Outputs {
		allOutputs = append(allOutputs, emptyWallet.BuildAnnotatedOutput(newTx, i))
	}

	// Get input assetid and asset amount
	for _, inp := range allInputs {
		log.WithField("Id", inp.AssetID.String()).Info("Input Asset Id")
		log.WithField("Amount", inp.Amount).Info("Input Asset Amount")
		log.WithField("Address", inp.Address).Info("Input Address")
	}
	// Get output assetid and asset amount
	for _, outp := range allOutputs {
		log.WithField("Id", outp.AssetID.String()).Info("Output Asset Id")
		log.WithField("Amount", outp.Amount).Info("Output Asset Amount")
		log.WithField("Address", outp.Address).Info("Output Address")
	}
	return allInputs, allOutputs
}

// Callback Calls the registered url with data about the transaction
func Callback(url string, callbackData *CallbackData) {
	trSkipVerify := &http.Transport{
		TLSClientConfig: &tls.Config{
			MaxVersion:         tls.VersionTLS11,
			InsecureSkipVerify: true,
		},
	}
	var netClient = &http.Client{
		Timeout:   time.Second * 30,
		Transport: trSkipVerify,
	}
	bytesRepresentation, _ := json.Marshal(callbackData)
	req, _ := http.NewRequest("POST", url, bytes.NewBuffer(bytesRepresentation))
	req.Header.Set("Referer", "Bytom Node")
	req.Header.Set("Content-Type", "application/json")
	// The client just ignores any response from the server
	_, errCall := netClient.Do(req)
	if errCall != nil {
		log.WithField("error", errCall).Info("Address Callback failed")
	}

}
