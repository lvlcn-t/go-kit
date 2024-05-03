package executors

import (
	"context"
	"errors"
	"fmt"
)

// Protector returns an effector that recovers from panics and returns them as errors.
func Protector(effector Effector) Effector {
	if effector == nil {
		return noopEffector
	}

	return func(ctx context.Context) (err error) {
		defer func() {
			if r := recover(); r != nil {
				pErr, ok := r.(error)
				if !ok {
					pErr = fmt.Errorf("%v", r)
				}
				err = errors.Join(err, pErr)
			}
		}()

		err = effector(ctx)
		return
	}
}
