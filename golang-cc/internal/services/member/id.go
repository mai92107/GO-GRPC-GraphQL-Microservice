package member

import "github.com/rafa/golang-cc/internal/utils/secure"

func newID() string { return secure.UUID() }
