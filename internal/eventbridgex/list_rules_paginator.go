package eventbridgex

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/eventbridge"
	"github.com/aws/aws-sdk-go/aws"
)

/*
   implemented the eventbridgex.ListRules one by referring to the "github.com/aws/aws-sdk-go-v2/service/quicksight".ListAnalyses paginator.

   The original, original code is here; https://github.com/aws/aws-sdk-go-v2/blob/service/quicksight/v1.18.0/service/quicksight/api_op_ListAnalyses.go#L158
   The license for the original code is here.; https://github.com/aws/aws-sdk-go-v2/blob/service/quicksight/v1.18.0/LICENSE.txt
*/

// ListRulesAPIClient is a client that implements the ListRules operation.
type ListRulesAPIClient interface {
	ListRules(context.Context, *eventbridge.ListRulesInput, ...func(*eventbridge.Options)) (*eventbridge.ListRulesOutput, error)
}

// ListRulesPaginatorOptions is the paginator options for ListRules
type ListRulesPaginatorOptions struct {
	// The maximum number of results to return.
	Limit int32

	// Set to true if pagination should stop if the service returns a pagination token
	// that matches the most recent token provided to the service.
	StopOnDuplicateToken bool
}

// ListRulesPaginator is a paginator for ListRules
type ListRulesPaginator struct {
	options   ListRulesPaginatorOptions
	client    ListRulesAPIClient
	params    *eventbridge.ListRulesInput
	nextToken *string
	firstPage bool
}

// NewListRulesPaginator returns a new ListRulesPaginator
func NewListRulesPaginator(client ListRulesAPIClient, params *eventbridge.ListRulesInput, optFns ...func(*ListRulesPaginatorOptions)) *ListRulesPaginator {
	if params == nil {
		params = &eventbridge.ListRulesInput{}
	}

	options := ListRulesPaginatorOptions{
		Limit: 100,
	}
	if params.Limit != nil {
		options.Limit = *params.Limit
	}

	for _, fn := range optFns {
		fn(&options)
	}

	return &ListRulesPaginator{
		options:   options,
		client:    client,
		params:    params,
		firstPage: true,
		nextToken: params.NextToken,
	}
}

// HasMorePages returns a boolean indicating whether more pages are available
func (p *ListRulesPaginator) HasMorePages() bool {
	return p.firstPage || (p.nextToken != nil && len(*p.nextToken) != 0)
}

// NextPage retrieves the next ListRules page.
func (p *ListRulesPaginator) NextPage(ctx context.Context, optFns ...func(*eventbridge.Options)) (*eventbridge.ListRulesOutput, error) {
	if !p.HasMorePages() {
		return nil, fmt.Errorf("no more pages available")
	}

	params := *p.params
	params.NextToken = p.nextToken

	params.Limit = aws.Int32(p.options.Limit)

	result, err := p.client.ListRules(ctx, &params, optFns...)
	if err != nil {
		return nil, err
	}
	p.firstPage = false

	prevToken := p.nextToken
	p.nextToken = result.NextToken

	if p.options.StopOnDuplicateToken &&
		prevToken != nil &&
		p.nextToken != nil &&
		*prevToken == *p.nextToken {
		p.nextToken = nil
	}

	return result, nil
}
