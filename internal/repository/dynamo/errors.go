package dynamo

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/adell/cloudops-release-intelligence/internal/domain"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/aws/smithy-go"
)

func wrapErr(op string, err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		return err
	}
	var cond *types.ConditionalCheckFailedException
	if errors.As(err, &cond) {
		return domain.AlreadyExistsError{Resource: "item", ID: op}
	}
	var tx *types.TransactionCanceledException
	if errors.As(err, &tx) {
		if isConditionalCancel(tx) {
			return domain.AlreadyExistsError{Resource: "item", ID: op}
		}
	}
	return fmt.Errorf("storage %s failed", op)
}

func isConditionalCancel(tx *types.TransactionCanceledException) bool {
	if tx == nil {
		return false
	}
	for _, r := range tx.CancellationReasons {
		if r.Code != nil && strings.EqualFold(*r.Code, "ConditionalCheckFailed") {
			return true
		}
	}
	return false
}

func isConditional(err error) bool {
	var cond *types.ConditionalCheckFailedException
	if errors.As(err, &cond) {
		return true
	}
	var api smithy.APIError
	if errors.As(err, &api) && api.ErrorCode() == "ConditionalCheckFailedException" {
		return true
	}
	return false
}
