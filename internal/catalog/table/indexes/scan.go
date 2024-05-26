package indexes

type ScanDirection int

const (
	ScanDirectionUnknown ScanDirection = iota
	ScanDirectionForward
	ScanDirectionBackward
)
