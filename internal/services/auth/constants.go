package auth

import "crypto/rand"

type TokenType string

var RandReader = rand.Read
