package database

import (
	"context"

	"github.com/google/uuid"
)

type RealDeviceQuerier struct {
	q SqlTransactionalQuerier
}

func NewRealDevicequerier(q SqlTransactionalQuerier) DeviceQuerier {
	return &RealDeviceQuerier{
		q: q,
	}
}

func (rd *RealDeviceQuerier) GetDeviceInfoByUser(ctx context.Context, arg GetDeviceInfoByUserParams) (uuid.UUID, error) {
	return rd.q.GetDeviceInfoByUser(ctx, arg)
}

func (rd *RealDeviceQuerier) CreateDeviceInfo(ctx context.Context, arg CreateDeviceInfoParams) (uuid.UUID, error) {
	return rd.q.CreateDeviceInfo(ctx, arg)
}
