package starportcmd

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/AlecAivazis/survey/v2"
	"github.com/AlecAivazis/survey/v2/terminal"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	"github.com/spf13/cobra"
	"github.com/zhigui-projects/zeus-onestop/starport/pkg/cliquiz"
	"github.com/zhigui-projects/zeus-onestop/starport/pkg/spn"
	"github.com/zhigui-projects/zeus-onestop/starport/services/networkbuilder"
)

var (
	spnNodeAddress   string
	spnAPIAddress    string
	spnFaucetAddress string
)

// NewNetwork creates a new network command that holds some other sub commands
// related to creating a new network collaboratively.
func NewNetwork() *cobra.Command {
	c := &cobra.Command{
		Use:   "network",
		Short: "Create and start blockchains collaboratively",
		Args:  cobra.ExactArgs(1),
	}

	// configure flags.
	c.PersistentFlags().StringVar(&spnNodeAddress, "spn-node-address", "https://rpc.alpha.starport.network:443", "SPN node address")
	c.PersistentFlags().StringVar(&spnAPIAddress, "spn-api-address", "https://rest.alpha.starport.network", "SPN api address")
	c.PersistentFlags().StringVar(&spnFaucetAddress, "spn-faucet-address", "https://faucet.alpha.starport.network", "SPN Faucet address")

	// add sub commands.
	c.AddCommand(NewNetworkAccount())
	c.AddCommand(NewNetworkChain())
	c.AddCommand(NewNetworkProposal())
	return c
}

var spnclient *spn.Client

func newNetworkBuilder(options ...networkbuilder.Option) (*networkbuilder.Builder, error) {
	var spnoptions []spn.Option
	// use test keyring backend on Gitpod in order to prevent prompting for keyring
	// password. This happens because Gitpod uses containers.
	//
	// when not on Gitpod, OS keyring backend is used which only asks password once.
	if os.Getenv("GITPOD_WORKSPACE_ID") != "" {
		spnoptions = append(spnoptions, spn.Keyring(keyring.BackendTest))
	}
	// init spnclient only once on start in order to spnclient to
	// reuse unlocked keyring in the following steps.
	if spnclient == nil {
		var err error
		if spnclient, err = spn.New(spnNodeAddress, spnAPIAddress, spnFaucetAddress, spnoptions...); err != nil {
			return nil, err
		}
	}
	return networkbuilder.New(spnclient, options...)
}

// ensureSPNAccount ensures that an SPN account has ben set by interactively asking
// users to create, import or pick an account.
func ensureSPNAccount(b *networkbuilder.Builder) error {
	if _, err := b.AccountInUse(); err == nil {
		return nil
	}

	title := "Starport Network"

	printSection(fmt.Sprintf("Account on %s", title))
	fmt.Printf("To use %s you need an account.\nPlease, select an account or create a new one:\n\n", title)

	account, err := createSPNAccount(b, title)
	if err != nil {
		return err
	}

	return b.AccountUse(account.Name)
}

// createAccount interactively creates a Cosmos account in OS keyring or fs keyring depending
// on the system.
func createSPNAccount(b *networkbuilder.Builder, title string) (account spn.Account, err error) {
	accounts, err := accountNames(b)
	if err != nil {
		return account, err
	}
	var (
		createAccount = "Create a new account"
		importAccount = "Import an account from mnemonic"
	)
	list := append(accounts, createAccount, importAccount)
	var (
		qs = []*survey.Question{
			{
				Name: "account",
				Prompt: &survey.Select{
					Message: "Choose an account:",
					Options: list,
				},
			},
		}
		answers = struct {
			Account string `survey:"account"`
		}{}
	)
	if err = survey.Ask(qs, &answers); err != nil {
		if err == terminal.InterruptErr {
			return account, context.Canceled
		}
		return account, err
	}

	switch answers.Account {
	case createAccount:
		var name string
		if err := cliquiz.Ask(cliquiz.NewQuestion("Account name", &name, cliquiz.Required())); err != nil {
			return account, err
		}

		if account, err = b.AccountCreate(name, ""); err != nil {
			return account, err
		}

		fmt.Printf("\n%s account has been created successfully!\nAccount address: %s \nMnemonic: %s\n\n",
			title,
			account.Address,
			account.Mnemonic,
		)

	case importAccount:
		var name string
		var mnemonic string
		if err := cliquiz.Ask(
			cliquiz.NewQuestion("Account name", &name, cliquiz.Required()),
			cliquiz.NewQuestion("Mnemonic", &mnemonic, cliquiz.Required()),
		); err != nil {
			return account, err
		}

		if account, err = b.AccountCreate(name, mnemonic); err != nil {
			return account, err
		}
		fmt.Printf("\n%s account has been imported successfully!\nAccount address: %s\n\n", title, account.Address)

	default:
		if account, err = b.AccountGet(answers.Account); err != nil {
			return account, err
		}
		fmt.Printf("\n%s account has been selected.\nAccount address: %s\n\n", title, account.Address)
	}

	return account, nil
}

func printSection(title string) {
	fmt.Printf("---------------------------------------------\n%s\n---------------------------------------------\n\n", title)
}

// accountNames retrieves a name list of accounts in the OS keyring.
func accountNames(b *networkbuilder.Builder) ([]string, error) {
	var names []string
	accounts, err := b.AccountList()
	if err != nil {
		return nil, err
	}
	for _, account := range accounts {
		names = append(names, account.Name)
	}
	return names, nil
}

func ensureSPNAccountHook(cmd *cobra.Command, args []string) error {
	nb, err := newNetworkBuilder()
	if err != nil {
		return err
	}
	err = ensureSPNAccount(nb)
	if err == context.Canceled {
		return errors.New("aborted")
	}
	return err
}
