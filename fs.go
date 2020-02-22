package dropbox

import (
	"os"
	"path/filepath"

	log "github.com/sirupsen/logrus"
)

func createFile(filePath string) (*os.File, error) {
	ensureDir(filePath)
	return os.Create(filePath)
}

func ensureDir(fileName string) {
	dirName := filepath.Dir(fileName)
	if _, serr := os.Stat(dirName); serr != nil {
		merr := os.MkdirAll(dirName, os.ModePerm)
		if merr != nil {
			log.WithError(merr).WithField("path", fileName).Error("error creating dir")
		}
	}
}
