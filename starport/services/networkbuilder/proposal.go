package networkbuilder

import (
	"context"

	"github.com/zhigui-projects/zeus-onestop/starport/pkg/spn"
)

// ProposalList lists proposals on a chain by status.
func (b *Builder) ProposalList(ctx context.Context, chainID string, options ...spn.ProposalListOption) ([]spn.Proposal, error) {
	account, err := b.AccountInUse()
	if err != nil {
		return nil, err
	}
	return b.spnclient.ProposalList(ctx, account.Name, chainID, options...)
}

// ProposalGet retrieves a proposal on a chain by id.
func (b *Builder) ProposalGet(ctx context.Context, chainID string, id int) (spn.Proposal, error) {
	account, err := b.AccountInUse()
	if err != nil {
		return spn.Proposal{}, err
	}
	return b.spnclient.ProposalGet(ctx, account.Name, chainID, id)
}

// Propose proposes given proposals in batch for chainID by using SPN accountName.
func (b *Builder) Propose(ctx context.Context, chainID string, proposals ...spn.ProposalOption) error {
	acc, err := b.AccountInUse()
	if err != nil {
		return err
	}
	return b.spnclient.Propose(ctx, acc.Name, chainID, proposals...)
}

// SubmitReviewals submits reviewals for proposals in batch for chainID by using SPN accountName.
func (b *Builder) SubmitReviewals(ctx context.Context, chainID string, reviewals ...spn.Reviewal) (gas uint64, broadcast func() error, err error) {
	acc, err := b.AccountInUse()
	if err != nil {
		return 0, nil, err
	}
	return b.spnclient.SubmitReviewals(ctx, acc.Name, chainID, reviewals...)
}
