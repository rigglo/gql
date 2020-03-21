package schema

import (
	"context"
)

type IntrospectionRule func(context.Context) bool
