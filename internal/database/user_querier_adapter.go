package database

import "context"

type RealUserQuerier struct {
	q *Queries
}

func NewRealUserQuerier(q *Queries) UserQuerier {
	return &RealUserQuerier{
		q: q,
	}
}

func (ru *RealUserQuerier) CreateUser(ctx context.Context, params CreateUserParams) error {
	return ru.q.CreateUser(ctx, params)
}

func (ru *RealUserQuerier) GetUserByEmail(ctx context.Context, email string) (GetUserByEmailRow, error) {
	return ru.q.GetUserByEmail(ctx, email)
}
