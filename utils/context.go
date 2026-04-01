package utils

import (
	"context"
	"orange-agent/domain"
	"orange-agent/repository/resource"
	"orange-agent/utils/logger"
)

type contextKey string

const userContextKey contextKey = "user"

// WithUser 将用户信息存入上下文
func WithUser(ctx context.Context, user *domain.User) context.Context {
	return context.WithValue(ctx, userContextKey, user)
}

// GetUserFromContext 从上下文中获取用户信息
func GetUserFromContext(ctx context.Context) (*domain.User, bool) {
	user, ok := ctx.Value(userContextKey).(*domain.User)
	return user, ok
}

// GetUserFromContextOrDefault 从上下文中获取用户信息，如果不存在则返回默认用户
func GetUserFromContextOrDefault(ctx context.Context) *domain.User {
	if user, ok := GetUserFromContext(ctx); ok {
		return user
	}
	repo := resource.GetRepositories()
	user, err := repo.User.GetUserById(1)
	if err != nil {
		logger.GetLogger().Error("GetUserFromContextOrDefault: %v", err)
	}
	return user
}
