package starportcmd

import (
	"errors"
	"fmt"

	"github.com/manifoldco/promptui"

	"github.com/spf13/cobra"
	"github.com/zhigui-projects/zeus-onestop/starport/pkg/clispinner"
	"github.com/zhigui-projects/zeus-onestop/starport/pkg/numbers"
	"github.com/zhigui-projects/zeus-onestop/starport/pkg/spn"
)

// NewNetworkProposalReject creates a new reject approve command to reject proposals
// for a chain.
func NewNetworkProposalReject() *cobra.Command {
	c := &cobra.Command{
		Use:   "reject [chain-id] [number<,...>]",
		Short: "Reject proposals",
		RunE:  networkProposalRejectHandler,
		Args:  cobra.ExactArgs(2),
	}
	return c
}

func networkProposalRejectHandler(cmd *cobra.Command, args []string) error {
	s := clispinner.New()
	defer s.Stop()

	s.SetText("Calculating gas...")

	var (
		chainID      = args[0]
		proposalList = args[1]
	)

	nb, err := newNetworkBuilder()
	if err != nil {
		return err
	}

	var reviewals []spn.Reviewal

	ids, err := numbers.ParseList(proposalList)
	if err != nil {
		return err
	}
	for _, id := range ids {
		reviewals = append(reviewals, spn.RejectProposal(id))
	}

	gas, broadcast, err := nb.SubmitReviewals(cmd.Context(), chainID, reviewals...)
	if err != nil {
		return err
	}

	s.Stop()

	// Prompt for confirmation
	prompt := promptui.Prompt{
		Label: fmt.Sprintf("This operation will cost about %v gas. Confirm the transaction",
			gas,
		),
		IsConfirm: true,
	}
	if _, err := prompt.Run(); err != nil {
		return errors.New("transaction aborted")
	}

	s.SetText("Rejecting...")
	s.Start()

	// Broadcast the transaction
	if err := broadcast(); err != nil {
		return err
	}
	s.Stop()

	fmt.Printf("Proposal(s) %s rejected ⛔️\n", numbers.List(ids, "#"))
	return nil
}
