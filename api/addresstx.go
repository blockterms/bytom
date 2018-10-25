package api

import (
	"context"

	log "github.com/sirupsen/logrus"
)

// POST /add-address-callback
func (a *API) addAddressCallback(ctx context.Context, ins struct {
	Address string `json:"address"`
	URL     string `json:"url"`
}) Response {
	isAdded, err := a.callbackStore.Add(ins.Address, ins.URL)
	if err != nil {
		return NewErrorResponse(err)
	}
	return NewSuccessResponse(isAdded)
}

// POST /list-address-callbacks
func (a *API) listAddressCallbacks(ctx context.Context, ins struct {
	Address string `json:"address"`
}) Response {
	callbacks, err := a.callbackStore.List(ins.Address)
	if err != nil {
		log.Errorf("listOfCallbacks: %v", err)
		return NewErrorResponse(err)
	}

	return NewSuccessResponse(callbacks)
}

// POST /remove-address-callback
func (a *API) removeAddressCallback(ctx context.Context, ins struct {
	Address string `json:"address"`
	URL     string `json:"url"`
}) Response {
	if err := a.callbackStore.Delete(ins.Address, ins.URL); err != nil {
		return NewErrorResponse(err)
	}
	return NewSuccessResponse(true)
}
