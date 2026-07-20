package commons

import "context"

func ContextSerialize(ctx context.Context, fields []string) map[string]any {
	m := map[string]any{}
	for _, name := range fields {
		vAny := ctx.Value(name)
		v, ok := vAny.(string)
		if ok {
			m[name] = v
		}
	}

	return m
}

func ContextFromMap(m map[string]any) context.Context {
	ctx := context.Background()

	for k, v := range m {
		ctx = context.WithValue(ctx, k, v)
	}

	return ctx
}