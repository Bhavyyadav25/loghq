package loghq

import "context"

type ctxFieldsKey struct{}

// ContextWithFields attaches logging fields to a context.
// These fields will be automatically included in log calls made with WithContext.
func ContextWithFields(ctx context.Context, fields ...Field) context.Context {
	existing := fieldsFromContext(ctx)
	merged := make([]Field, len(existing)+len(fields))
	copy(merged, existing)
	copy(merged[len(existing):], fields)
	return context.WithValue(ctx, ctxFieldsKey{}, merged)
}

// fieldsFromContext extracts logging fields from a context.
func fieldsFromContext(ctx context.Context) []Field {
	if ctx == nil {
		return nil
	}
	if f, ok := ctx.Value(ctxFieldsKey{}).([]Field); ok {
		return f
	}
	return nil
}
