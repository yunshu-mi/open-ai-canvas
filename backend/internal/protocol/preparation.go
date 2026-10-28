package protocol

import (
	"context"
	"fmt"
	"strings"
	"time"
)

type PreparationAdapter interface {
	BuildPreparation(context.Context, ManifestPreparation, GenerationRequest) (RequestSpec, error)
	ParsePreparation(context.Context, ManifestPreparation, GenerationRequest, []byte) (any, string, error)
}

func (a manifestAdapter) BuildPreparation(_ context.Context, step ManifestPreparation, request GenerationRequest) (RequestSpec, error) {
	return buildManifestOperation(step.Operation, a.manifest.Auth, request, "")
}

func (a manifestAdapter) ParsePreparation(_ context.Context, step ManifestPreparation, request GenerationRequest, body []byte) (any, string, error) {
	payload, err := decodeObject(body)
	if err != nil {
		return nil, "", err
	}
	env := map[string]any{"request": manifestRequestValues(request), "response": payload}
	if step.ExpiresAt != nil {
		value, err := evaluateManifestValue(step.ExpiresAt, env)
		if err != nil {
			return nil, "", err
		}
		expiresAt, err := time.Parse(time.RFC3339Nano, manifestString(value))
		if err != nil || !expiresAt.After(time.Now()) {
			return nil, "", fmt.Errorf("准备步骤 %s 的快照已过期或有效期无效", step.ID)
		}
	}
	for _, rule := range step.Validations {
		value, err := evaluateManifestValue(rule.Assert, env)
		if err != nil {
			return nil, "", err
		}
		if !manifestTruthy(value) {
			return nil, "", fmt.Errorf("%s", rule.Message)
		}
	}
	value, err := evaluateManifestValue(step.Value, env)
	if err != nil {
		return nil, "", err
	}
	if value == nil {
		return nil, "", fmt.Errorf("preparation %s returned no value", step.ID)
	}
	return value, strings.TrimSpace(manifestResponseString(step.Message, env)), nil
}
