package cosmosfaucet

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/zhigui-projects/zeus-onestop/starport/pkg/cosmoscoin"
	"github.com/zhigui-projects/zeus-onestop/starport/pkg/xhttp"
)

const (
	statusOK    = "ok"
	statusError = "error"
)

type transferRequest struct {
	// AccountAddress to request for coins.
	AccountAddress string `json:"address"`

	// Coins that are requested.
	// default ones used when this one isn't provided.
	Coins []string `json:"coins"`
}

type transferResponse struct {
	Error     string     `json:"error,omitempty"`
	Transfers []transfer `json:"transfers,omitempty"`
}

type transfer struct {
	Coin   string `json:"coin"`
	Status string `json:"status"`
	Error  string `json:"error,omitempty"`
}

func (f Faucet) faucetHandler(w http.ResponseWriter, r *http.Request) {
	var req transferRequest

	// decode request into req.
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		responseError(w, http.StatusBadRequest, err)
		return
	}

	// determine coins to transfer.
	coins, err := f.coinsToTransfer(req)
	if err != nil {
		responseError(w, http.StatusBadRequest, err)
		return
	}

	// send coins and create a transfers response.
	var transfers []transfer

	for _, coin := range coins {
		t := transfer{
			Coin:   coin.String(),
			Status: statusOK,
		}

		if err := f.Transfer(r.Context(), req.AccountAddress, coin.amount, coin.denom); err != nil {
			if err == context.Canceled {
				return
			}

			t.Status = statusError
			t.Error = err.Error()
		}

		transfers = append(transfers, t)
	}

	// send the response.
	responseSuccess(w, transfers)
}

// coinsToTransfer determines tokens to transfer from transfer request.
func (f Faucet) coinsToTransfer(req transferRequest) ([]coin, error) {
	if len(req.Coins) == 0 {
		return f.coins, nil
	}

	var coins []coin
	for _, c := range req.Coins {
		amount, denom, err := cosmoscoin.Parse(c)
		if err != nil {
			return nil, err
		}
		coins = append(coins, coin{amount, denom})
	}

	return coins, nil
}

func responseSuccess(w http.ResponseWriter, transfers []transfer) {
	xhttp.ResponseJSON(w, http.StatusOK, transferResponse{
		Transfers: transfers,
	})
}

func responseError(w http.ResponseWriter, code int, err error) {
	xhttp.ResponseJSON(w, code, transferResponse{
		Error: err.Error(),
	})
}
