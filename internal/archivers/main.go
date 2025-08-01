package archivers

type (
	ArchiverInterface interface {
		Archivate(sourceDir, targetFile string, excludeDirs []string) (string, error)
	}

	Archiver struct {
	}
)

func NewArchiver() *Archiver {

	return &Archiver{}
}

func (d *Archiver) LoadAllowedDrivers() map[string]func() ArchiverInterface {

	return map[string]func() ArchiverInterface{
		"tar": NewTar,
		"zip": NewZip,
	}
}
