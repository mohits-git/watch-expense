package authctx

import "context"

type tokenKeyType string

const tokenKey tokenKeyType = "token"

func WithToken(ctx context.Context, token string) context.Context {
	return context.WithValue(ctx, tokenKey, token)
}

func TokenFromCtx(ctx context.Context) (string, bool) {
	val := ctx.Value(tokenKey)
	if val == nil || val == "" {
		return "", false
	}

	token, ok := val.(string)
	return token, ok
}
