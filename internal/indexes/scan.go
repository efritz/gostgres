package indexes

type tidScanner interface {
	Scan() (int, error)
}

type tidScannerFunc func() (int, error)

func (f tidScannerFunc) Scan() (int, error) {
	return f()
}

type ScanDirection int

const (
	ScanDirectionUnknown ScanDirection = iota
	ScanDirectionForward
	ScanDirectionBackward
)
