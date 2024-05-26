package scan

type TIDScanner interface {
	Scan() (int64, error)
}

type TIDScannerFunc func() (int64, error)

func (f TIDScannerFunc) Scan() (int64, error) {
	return f()
}
