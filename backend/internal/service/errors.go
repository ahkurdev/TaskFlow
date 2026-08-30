package service

import "errors"

// ErrNotFound is returned when a resource doesn't exist, or exists but
// doesn't belong to the requesting user (we don't distinguish the two to
// avoid leaking existence of other users' resources).
var ErrNotFound = errors.New("resource not found")
