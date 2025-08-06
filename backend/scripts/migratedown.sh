#!/bin/bash

if [ -f .env ]; then
    source .env
fi

cd sql/schema
goose turso $TURSO_DATABASE_URL down