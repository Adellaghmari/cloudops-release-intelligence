package dynamo

import (
	"context"
	"strings"
	"sync"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

// Fake implements the DynamoDB operations this adapter uses. It exists because
// this machine has neither Docker nor Java, so DynamoDB Local / Testcontainers
// cannot run. It is a protocol-compatible subset, not Amazon DynamoDB.
type Fake struct {
	mu     sync.Mutex
	tables map[string]*fakeTable
}

type fakeTable struct {
	items map[string]map[string]map[string]types.AttributeValue // pk -> sk -> item
}

func NewFake() *Fake {
	return &Fake{tables: map[string]*fakeTable{}}
}

func (f *Fake) CreateTable(_ context.Context, in *dynamodb.CreateTableInput, _ ...func(*dynamodb.Options)) (*dynamodb.CreateTableOutput, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	name := aws.ToString(in.TableName)
	if _, ok := f.tables[name]; !ok {
		f.tables[name] = &fakeTable{items: map[string]map[string]map[string]types.AttributeValue{}}
	}
	return &dynamodb.CreateTableOutput{}, nil
}

func (f *Fake) DescribeTable(ctx context.Context, in *dynamodb.DescribeTableInput, _ ...func(*dynamodb.Options)) (*dynamodb.DescribeTableOutput, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	name := aws.ToString(in.TableName)
	if _, ok := f.tables[name]; !ok {
		return nil, &types.ResourceNotFoundException{Message: aws.String("table not found")}
	}
	return &dynamodb.DescribeTableOutput{Table: &types.TableDescription{TableName: in.TableName}}, nil
}

func (f *Fake) PutItem(ctx context.Context, in *dynamodb.PutItemInput, _ ...func(*dynamodb.Options)) (*dynamodb.PutItemOutput, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	tbl := f.table(aws.ToString(in.TableName))
	pk, sk := strAttr(in.Item["PK"]), strAttr(in.Item["SK"])
	if in.ConditionExpression != nil && strings.Contains(*in.ConditionExpression, "attribute_not_exists(PK)") {
		if _, ok := tbl.items[pk][sk]; ok {
			return nil, &types.ConditionalCheckFailedException{Message: aws.String("exists")}
		}
	}
	f.putLocked(tbl, in.Item)
	return &dynamodb.PutItemOutput{}, nil
}

func (f *Fake) GetItem(ctx context.Context, in *dynamodb.GetItemInput, _ ...func(*dynamodb.Options)) (*dynamodb.GetItemOutput, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	tbl := f.table(aws.ToString(in.TableName))
	pk, sk := strAttr(in.Key["PK"]), strAttr(in.Key["SK"])
	item, ok := tbl.items[pk][sk]
	if !ok {
		return &dynamodb.GetItemOutput{}, nil
	}
	return &dynamodb.GetItemOutput{Item: cloneItem(item)}, nil
}

func (f *Fake) Query(ctx context.Context, in *dynamodb.QueryInput, _ ...func(*dynamodb.Options)) (*dynamodb.QueryOutput, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	tbl := f.table(aws.ToString(in.TableName))
	expr := aws.ToString(in.KeyConditionExpression)
	vals := in.ExpressionAttributeValues
	var items []map[string]types.AttributeValue
	index := aws.ToString(in.IndexName)
	switch {
	case index == "" && strings.Contains(expr, "begins_with(SK"):
		pk := strAttr(vals[":pk"])
		prefix := strAttr(vals[":sk"])
		for sk, item := range tbl.items[pk] {
			if strings.HasPrefix(sk, prefix) {
				items = append(items, cloneItem(item))
			}
		}
	case index == "GSI1":
		pk := strAttr(vals[":pk"])
		prefix := ""
		if v, ok := vals[":sk"]; ok {
			prefix = strAttr(v)
		}
		for _, bySK := range tbl.items {
			for _, item := range bySK {
				if strAttr(item["GSI1PK"]) != pk {
					continue
				}
				if prefix == "" || strings.HasPrefix(strAttr(item["GSI1SK"]), prefix) {
					items = append(items, cloneItem(item))
				}
			}
		}
	case index == "GSI2":
		pk := strAttr(vals[":pk"])
		for _, bySK := range tbl.items {
			for _, item := range bySK {
				if strAttr(item["GSI2PK"]) == pk {
					items = append(items, cloneItem(item))
				}
			}
		}
	default:
		pk := strAttr(vals[":pk"])
		for _, item := range tbl.items[pk] {
			items = append(items, cloneItem(item))
		}
	}
	if in.ScanIndexForward != nil && !*in.ScanIndexForward {
		// sort handled by adapter after decode; leave unsorted
	}
	return &dynamodb.QueryOutput{Items: items}, nil
}

func (f *Fake) DeleteItem(ctx context.Context, in *dynamodb.DeleteItemInput, _ ...func(*dynamodb.Options)) (*dynamodb.DeleteItemOutput, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	tbl := f.table(aws.ToString(in.TableName))
	pk, sk := strAttr(in.Key["PK"]), strAttr(in.Key["SK"])
	if bySK, ok := tbl.items[pk]; ok {
		delete(bySK, sk)
	}
	return &dynamodb.DeleteItemOutput{}, nil
}

func (f *Fake) TransactWriteItems(ctx context.Context, in *dynamodb.TransactWriteItemsInput, _ ...func(*dynamodb.Options)) (*dynamodb.TransactWriteItemsOutput, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	type pending struct {
		table string
		item  map[string]types.AttributeValue
	}
	var writes []pending
	for _, op := range in.TransactItems {
		if op.Put == nil {
			continue
		}
		tbl := f.table(aws.ToString(op.Put.TableName))
		pk, sk := strAttr(op.Put.Item["PK"]), strAttr(op.Put.Item["SK"])
		if op.Put.ConditionExpression != nil && strings.Contains(*op.Put.ConditionExpression, "attribute_not_exists(PK)") {
			if _, ok := tbl.items[pk][sk]; ok {
				return nil, &types.ConditionalCheckFailedException{Message: aws.String("exists")}
			}
		}
		writes = append(writes, pending{table: aws.ToString(op.Put.TableName), item: op.Put.Item})
	}
	for _, w := range writes {
		f.putLocked(f.table(w.table), w.item)
	}
	return &dynamodb.TransactWriteItemsOutput{}, nil
}

func (f *Fake) table(name string) *fakeTable {
	t, ok := f.tables[name]
	if !ok {
		t = &fakeTable{items: map[string]map[string]map[string]types.AttributeValue{}}
		f.tables[name] = t
	}
	return t
}

func (f *Fake) putLocked(tbl *fakeTable, item map[string]types.AttributeValue) {
	pk, sk := strAttr(item["PK"]), strAttr(item["SK"])
	if tbl.items[pk] == nil {
		tbl.items[pk] = map[string]map[string]types.AttributeValue{}
	}
	tbl.items[pk][sk] = cloneItem(item)
}

func strAttr(v types.AttributeValue) string {
	if v == nil {
		return ""
	}
	s, ok := v.(*types.AttributeValueMemberS)
	if !ok {
		return ""
	}
	return s.Value
}

func cloneItem(in map[string]types.AttributeValue) map[string]types.AttributeValue {
	out := make(map[string]types.AttributeValue, len(in))
	for k, v := range in {
		out[k] = v
	}
	return out
}
