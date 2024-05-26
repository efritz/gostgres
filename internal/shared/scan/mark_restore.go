package scan

type MarkRestorer interface {
	Mark()
	Restore()
}
