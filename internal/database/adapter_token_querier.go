package database

import (
	"context"
)

type RealTokenQuerier struct {
	q SqlTransactionalQuerier
}

func NewRealTokenQuerier(q SqlTransactionalQuerier) TokenQuerier {
	return &RealTokenQuerier{
		q: q,
	}
}

func (rt *RealTokenQuerier) CreateRefreshToken(ctx context.Context, arg CreateRefreshTokenParams) error {
	return rt.q.CreateRefreshToken(ctx, arg)
}

func (rt *RealTokenQuerier) RevokeToken(ctx context.Context, arg RevokeTokenParams) error {
	return rt.q.RevokeToken(ctx, arg)
}

func (rt *RealTokenQuerier) GetRefreshByUserAndDevice(ctx context.Context, arg GetRefreshByUserAndDeviceParams) (RefreshToken, error) {
	return rt.q.GetRefreshByUserAndDevice(ctx, arg)
}
