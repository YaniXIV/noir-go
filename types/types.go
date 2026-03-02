package types

// Visibility indicates whether a witness is public or private.
type Visibility string

const (
	// VisibilityPublic marks a witness as public.
	VisibilityPublic  Visibility = "public"
	// VisibilityPrivate marks a witness as private.
	VisibilityPrivate Visibility = "private"
)
