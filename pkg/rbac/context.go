package rbac

import (
	"context"
)

type contextKey string

const (
	userContextKey      contextKey = "rbac_user"
	namespaceContextKey contextKey = "rbac_namespace"
)

// UserFromContext retrieves the current user from context
func UserFromContext(ctx context.Context) (*User, bool) {
	user, ok := ctx.Value(userContextKey).(*User)
	return user, ok
}

// ContextWithUser adds a user to the context
func ContextWithUser(ctx context.Context, user *User) context.Context {
	return context.WithValue(ctx, userContextKey, user)
}

// NamespaceFromContext retrieves the current namespace from context
func NamespaceFromContext(ctx context.Context) (*Namespace, bool) {
	ns, ok := ctx.Value(namespaceContextKey).(*Namespace)
	return ns, ok
}

// ContextWithNamespace adds a namespace to the context
func ContextWithNamespace(ctx context.Context, namespace *Namespace) context.Context {
	return context.WithValue(ctx, namespaceContextKey, namespace)
}
