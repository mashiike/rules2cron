package rules2cron

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/eventbridge"
	"github.com/aws/aws-sdk-go-v2/service/eventbridge/types"
	"github.com/mashiike/rules2cron/internal/eventbridgex"
)

type App struct {
	client    *eventbridge.Client
	converter *Converter
}

func New(ctx context.Context, converter *Converter) (*App, error) {
	opts := make([]func(*config.LoadOptions) error, 0)

	if region := os.Getenv("AWS_DEFAULT_REGION"); region != "" {
		opts = append(opts, config.WithRegion(region))
	}
	if endpoint := os.Getenv("EVENTBRIDGE_ENDPOINT"); endpoint != "" {
		opts = append(opts, config.WithEndpointResolverWithOptions(
			aws.EndpointResolverWithOptionsFunc(func(service, region string, _ ...interface{}) (aws.Endpoint, error) {
				if service == eventbridge.ServiceID {
					return aws.Endpoint{
						URL:           endpoint,
						PartitionID:   "aws",
						SigningRegion: region,
					}, nil
				}
				return aws.Endpoint{}, &aws.EndpointNotFoundError{}
			})))
	}
	awsCfg, err := config.LoadDefaultConfig(ctx, opts...)
	if err != nil {
		return nil, err
	}

	app := &App{
		client:    eventbridge.NewFromConfig(awsCfg),
		converter: converter,
	}
	return app, err
}

func (app *App) Run(w io.Writer, showDisabled bool) error {
	return app.RunWithContext(context.Background(), w, showDisabled)
}

func (app *App) RunWithContext(ctx context.Context, w io.Writer, showDisabled bool) error {
	p := eventbridgex.NewListRulesPaginator(app.client, &eventbridge.ListRulesInput{})
	for p.HasMorePages() {
		output, err := p.NextPage(ctx)
		if err != nil {
			return err
		}
		for _, rule := range output.Rules {
			if rule.ScheduleExpression == nil {
				log.Printf("[debug] rule %s is not scheduled rule, skip", *rule.Arn)
				continue
			}
			if !showDisabled && rule.State == types.RuleStateDisabled {
				log.Printf("[debug] rule %s is disabled, skip", *rule.Arn)
				continue
			}
			cronExpression, err := app.converter.Convert(*rule.ScheduleExpression)
			if err != nil {
				log.Printf("[warn] rule %s: %s", *rule.Name, err.Error())
				continue
			}
			fmt.Fprintf(w, "%s\t%s\n", cronExpression, *rule.Name)
		}
	}
	return nil
}
