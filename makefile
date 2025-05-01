.PHONY: all generate postprocess move rename qualify mocks

all: generate postprocess move rename qualify mocks

generate:
	sqlc generate

postprocess: generate
	sqlc-qol add-nosec \
	"internal/database/*.sql.go" \
	--csv=./data/targets.csv

move:
	mv ./internal/database/models.go ./internal/models/db.go

rename: move
	sed -i "s/^package database$$/package models/" internal/models/db.go

qualify: rename move
	sqlc-qol qualify-models \
	-m internal/models/db.go \
	-d internal/database \
	-i github.com/seanhuebl/unity-wealth/internal/models
	goimports -w internal/database

mocks: qualify
	mockery --config mockery_database.yaml
	mockery --config mockery_auth.yaml