package loader

import "github.com/efritz/gostgres/internal/shared"

type FieldDescription struct {
	Name     string
	TypeKind shared.TypeKind
}
