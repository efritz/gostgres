package indexes

type tidScanner interface {
	Scan() (int64, error)
}

type tidScannerFunc func() (int64, error)

func (f tidScannerFunc) Scan() (int64, error) {
	return f()
}

type ScanDirection int

const (
	ScanDirectionUnknown ScanDirection = iota
	ScanDirectionForward
	ScanDirectionBackward
)
