package cosmosfaucet

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	chaincmdrunner "github.com/zhigui-projects/zeus-onestop/starport/pkg/chaincmd/runner"
)

// TotalTransferredAmount returns the total transferred amount from faucet account to toAccountAddress.
func (f Faucet) TotalTransferredAmount(ctx context.Context, toAccountAddress, denom string) (amount uint64, err error) {
	fromAccount, err := f.runner.ShowAccount(ctx, f.accountName)
	if err != nil {
		return 0, err
	}

	events, err := f.runner.QueryTxEvents(ctx,
		chaincmdrunner.NewEventSelector("message", "sender", fromAccount.Address),
		chaincmdrunner.NewEventSelector("transfer", "recipient", toAccountAddress))
	if err != nil {
		return 0, err
	}

	for _, event := range events {
		if event.Type == "transfer" {
			for _, attr := range event.Attributes {
				if attr.Key == "amount" {
					if !strings.HasSuffix(attr.Value, denom) {
						continue
					}

					amountStr := strings.TrimRight(attr.Value, denom)
					if a, err := strconv.ParseUint(amountStr, 10, 64); err == nil {
						amount += a
					}
				}
			}
		}
	}

	return amount, nil
}

// Transfer transfer amount of tokens from the faucet account to toAccountAddress.
func (f Faucet) Transfer(ctx context.Context, toAccountAddress string, amount uint64, denom string) error {
	amountStr := fmt.Sprintf("%d%s", amount, denom)

	totalSent, err := f.TotalTransferredAmount(ctx, toAccountAddress, denom)
	if err != nil {
		return err
	}

	if f.coinsMax[denom] != 0 && totalSent >= f.coinsMax[denom] {
		return fmt.Errorf("account has reached maximum credit allowed per account (%d)", f.coinsMax[denom])
	}

	fromAccount, err := f.runner.ShowAccount(ctx, f.accountName)
	if err != nil {
		return err
	}

	return f.runner.BankSend(ctx, fromAccount.Address, toAccountAddress, amountStr)
}
