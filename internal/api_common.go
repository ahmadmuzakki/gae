package internal

import "context"

var namespace = "key that holds the namespace"

func GetNamespace(ctx context.Context) string {
	if ns, ok := ctx.Value(&namespace).(string); ok {
		return ns
	}
	return ""
}

func WithNamespace(ctx context.Context, ns string) context.Context {
	return context.WithValue(ctx, &namespace, ns)
}
