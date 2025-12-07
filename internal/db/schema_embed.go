package db

import _ "embed"

// Schema holds the embedded SQL schema.
//
//go:embed schema.sql
var Schema string
