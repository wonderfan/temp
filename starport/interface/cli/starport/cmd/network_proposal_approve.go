package starportcmd

import (
	"errors"
	"fmt"
	"io/ioutil"

	"github.com/manifoldco/promptui"

	"github.com/spf13/cobra"
	"github.com/zhigui-projects/zeus-onestop/starport/pkg/clispinner"
	"github.com/zhigui-projects/zeus-onestop/starport/pkg/numbers"
	"github.com/zhigui-projects/zeus-onestop/starport/pkg/spn"
)

const (
	flagNoVerification = "no-verification"
)

// NewNetworkProposalApprove creates a new proposal approve command to approve proposals
// for a chain.
func NewNetworkProposalApprove() *cobra.Command {
	c := &cobra.Command{
		Use:     "approve [chain-id] [number<,...>]",
		Aliases: []string{"accept"},
		Short:   "Approve proposals",
		RunE:    networkProposalApproveHandler,
		Args:    cobra.ExactArgs(2),
	}
	c.Flags().AddFlagSet(flagSetHomes())
	c.Flags().Bool(flagNoVerification, false, "approve the proposals without verifying them")
	return c
}

func networkProposalApproveHandler(cmd *cobra.Command, args []string) error {
	s := clispinner.New()
	s.Stop()
	defer s.Stop()

	var (
		chainID      = args[0]
		proposalList = args[1]
	)

	nb, err := newNetworkBuilder()
	if err != nil {
		return err
	}

	var reviewals []spn.Reviewal

	// Get the list of proposal ids
	ids, err := numbers.ParseList(proposalList)
	if err != nil {
		return err
	}

	// Verify the proposals are valid
	noVerification, err := cmd.Flags().GetBool(flagNoVerification)
	if err != nil {
		return err
	}
	if !noVerification {
		// Verify the proposal
		// This operation generate the genesis in a temporary directory and verify this genesis is valid
		s.SetText("Verifying proposals...")
		s.Start()

		// Check if custom home is provided
		home, _, err := getHomeFlags(cmd)
		if err != nil {
			return err
		}

		verified, err := nb.VerifyProposals(cmd.Context(), chainID, home, ids, ioutil.Discard)
		if err != nil {
			return err
		}
		if !verified {
			return fmt.Errorf("genesis from proposal(s) %s is invalid", numbers.List(ids, "#"))
		}
		s.Stop()
	}

	s.SetText("Calculating gas...")
	s.Start()

	// Submit the approve reviewals
	for _, id := range ids {
		reviewals = append(reviewals, spn.ApproveProposal(id))
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

	s.SetText("Approving...")
	s.Start()

	// Broadcast the transaction
	if err := broadcast(); err != nil {
		return err
	}
	s.Stop()

	fmt.Printf("Proposal(s) %s approved ✅\n", numbers.List(ids, "#"))
	return nil
}
