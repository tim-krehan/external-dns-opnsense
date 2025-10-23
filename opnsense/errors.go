package opnsense

import (
	"errors"
)

var ErrFailedToApply = errors.New("failed to apply changes on opnsense")
var ErrFailedToCreate = errors.New("failed to create dns entry")
var ErrFailedToUpdate = errors.New("failed to update dns entry")
var ErrFailedToDelete = errors.New("failed to delete dns entry")
var ErrApiReturnedError = errors.New("api returned an error")
