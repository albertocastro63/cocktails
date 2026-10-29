package dynamo

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

// scanAll runs a DynamoDB Scan to completion, following LastEvaluatedKey until
// the table (or filtered set) is exhausted. A single Scan call returns at most
// 1 MB of data, so callers that need every matching item MUST page through the
// results; issuing one Scan silently truncates large tables. All store read
// paths funnel through here so that fix lives in one place.
func scanAll(ctx context.Context, client *dynamodb.Client, input *dynamodb.ScanInput) ([]map[string]types.AttributeValue, error) {
	var items []map[string]types.AttributeValue
	for {
		out, err := client.Scan(ctx, input)
		if err != nil {
			return nil, err
		}
		items = append(items, out.Items...)
		if len(out.LastEvaluatedKey) == 0 {
			return items, nil
		}
		input.ExclusiveStartKey = out.LastEvaluatedKey
	}
}
