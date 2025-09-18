package dynamo

import (
	"context"
	"maps"
	"slices"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/retry"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	zlog "github.com/rs/zerolog/log"

	"github.com/mfelipe/go-feijoada/stream-buffer/models"
	sccfg "github.com/mfelipe/go-feijoada/stream-consumer/config"
)

type client interface {
	BatchWriteItem(ctx context.Context, params *dynamodb.BatchWriteItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.BatchWriteItemOutput, error)
}

type Client struct {
	db        client
	tableName string
}

func New(cfg sccfg.DynamoDB) *Client {
	awsCfg, err := config.LoadDefaultConfig(context.Background())
	if err != nil {
		zlog.Fatal().Err(err).Msg("failed to load default AWS SDK config")
	}

	var options []func(*dynamodb.Options)

	//TODO: move retry and backoff options to Config
	options = append(options, func(o *dynamodb.Options) {
		o.Retryer = retry.NewStandard(func(o *retry.StandardOptions) {
			o.MaxAttempts = cfg.RetryMax
			o.MaxBackoff = cfg.RetryWaitMax
			o.Backoff = retry.NewExponentialJitterBackoff(cfg.RetryWaitMax)
		})
		o.RetryMaxAttempts = cfg.RetryMax
		o.RetryMode = aws.RetryModeAdaptive
	})

	dbClient := dynamodb.NewFromConfig(awsCfg, options...)

	return &Client{
		db:        dbClient,
		tableName: cfg.TableName,
	}
}

// BatchWrite write items in a batch into DynamoDB. Should always return the message ids that failed.
func (c *Client) BatchWrite(ctx context.Context, messages map[string]models.Message) ([]string, error) {
	var unpersisted = make([]string, 0)

	var items []types.WriteRequest
	for id, msg := range messages {
		item := c.messageToItem(id, msg)
		items = append(items, types.WriteRequest{
			PutRequest: &types.PutRequest{
				Item: item,
			},
		})
	}

	input := &dynamodb.BatchWriteItemInput{
		RequestItems: map[string][]types.WriteRequest{
			c.tableName: items,
		},
	}

	output, err := c.db.BatchWriteItem(ctx, input)
	if output != nil {
		if wrs, ok := output.UnprocessedItems[c.tableName]; ok {
			for _, wr := range wrs {
				//TODO: handle is safely
				if av, ok := wr.PutRequest.Item["id"]; ok {
					id := av.(*types.AttributeValueMemberS).Value
					unpersisted = append(unpersisted, id)
				}
				unpersisted = append(unpersisted)
			}
		}
	} else {
		unpersisted = slices.Collect(maps.Keys(messages))
	}

	return unpersisted, err
}

func (c *Client) messageToItem(id string, msg models.Message) map[string]types.AttributeValue {
	item := map[string]types.AttributeValue{
		"id": &types.AttributeValueMemberS{
			Value: id,
		},
		"timestamp": &types.AttributeValueMemberS{
			Value: msg.Timestamp.Format(time.RFC3339),
		},
		"origin": &types.AttributeValueMemberS{
			Value: msg.Origin,
		},
		"schemaURI": &types.AttributeValueMemberS{
			Value: msg.SchemaURI,
		},
		"data": &types.AttributeValueMemberS{
			Value: string(msg.Data),
		},
	}

	return item
}
