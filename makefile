.PHONY: all generate postprocess

all: generate postprocess

generate:
	sqlc generate

postprocess: generate
	sqlc-qol add-nosec \
	"internal/database/*.sql.go" \
	--csv=./data/targets.csv