package dropbox

import (
	"bytes"
	"io"

	"github.com/dropbox/dropbox-sdk-go-unofficial/dropbox"
	"github.com/dropbox/dropbox-sdk-go-unofficial/dropbox/files"
	"github.com/dropbox/dropbox-sdk-go-unofficial/dropbox/users"
	log "github.com/sirupsen/logrus"
)

func GetUser(token string) (*users.FullAccount, error) {
	config := dropbox.Config{
		Token:    token,
		LogLevel: dropbox.LogDebug,
	}
	client := users.New(config)
	account, err := client.GetCurrentAccount()
	if err != nil {
		log.WithError(err).Error("failed to get current user account")
		return nil, err
	}
	return account, nil
}

func GetUserRootFolders(token string) ([]*files.FolderMetadata, error) {
	config := dropbox.Config{
		Token:    token,
		LogLevel: dropbox.LogDebug,
	}
	dbx := files.New(config)
	folders := make([]*files.FolderMetadata, 0)
	args := files.ListFolderArg{
		Recursive: false,
	}
	listFolderResult, err := dbx.ListFolder(&args)
	if err != nil {
		log.WithError(err).Error("failed to list fodlers")
		return nil, err
	}
	for _, metadata := range listFolderResult.Entries {
		switch metadata.(type) {
		case *files.FolderMetadata:
			folderEntry, _ := metadata.(*files.FolderMetadata)
			log.WithField("folder", folderEntry.PathLower).Info("found root folder")
			folders = append(folders, folderEntry)
		default:
			log.Debug("ignore")
		}
	}
	return folders, nil

}

func CreateFolder(token, folderName string) error {
	config := dropbox.Config{
		Token:    token,
		LogLevel: dropbox.LogDebug,
	}
	dbx := files.New(config)
	args := files.CreateFolderArg{
		Path: folderName,
	}
	result, err := dbx.CreateFolderV2(&args)
	if err != nil {
		log.WithField("path", folderName).WithError(err).Error("failed to create folder")
		return err
	}
	log.WithField("result", result).Debug("created folder")
	return nil
}
func ReplaceFile(token, filePath string, content io.Reader) error {
	err := DeleteFile(token, filePath)
	if err != nil {
		log.WithError(err).Info("failed to delete file but ignore")
	}
	return CreateFile(token, filePath, content)
}

func DeleteFile(token, filePath string) error {
	config := dropbox.Config{
		Token:    token,
		LogLevel: dropbox.LogDebug,
	}
	dbx := files.New(config)
	args := files.DeleteArg{
		Path: filePath,
	}
	result, err := dbx.DeleteV2(&args)
	if err != nil {
		log.WithField("path", filePath).WithError(err).Error("failed to delete file")
		return err
	}
	log.WithField("result", result).Debug("deleted file")
	return nil
}

func CreateFile(token, filePath string, content io.Reader) error {
	config := dropbox.Config{
		Token:    token,
		LogLevel: dropbox.LogDebug,
	}
	dbx := files.New(config)
	args := files.NewCommitInfo(filePath)
	result, err := dbx.Upload(args, content)
	if err != nil {
		log.WithField("path", filePath).WithError(err).Error("failed to create file")
		return err
	}
	log.WithField("result", result).Debug("created file")
	return nil
}

func GetEntriesForPath(token, path string) ([]files.IsMetadata, error) {
	config := dropbox.Config{
		Token:    token,
		LogLevel: dropbox.LogDebug,
	}
	dbx := files.New(config)
	args := files.ListFolderArg{
		Path:      path,
		Recursive: true,
	}
	listFolderResult, err := dbx.ListFolder(&args)
	if err != nil {
		log.WithField("path", path).WithError(err).Error("failed to list folder")
		return []files.IsMetadata{}, err
	}
	return listFolderResult.Entries, nil
}

func FindFilesAndDirs(token, path string) ([]*files.FileMetadata, []*files.FolderMetadata, error) {
	entries, err := GetEntriesForPath(token, path)
	filesEntrs := make([]*files.FileMetadata, 0)
	dirsEntrs := make([]*files.FolderMetadata, 0)
	if err != nil {
		log.WithField("path", path).WithError(err).Error("failed to list folder")
		return filesEntrs, dirsEntrs, err
	}

	for _, metadata := range entries {
		switch metadata.(type) {
		case *files.FileMetadata:
			fileEntry, _ := metadata.(*files.FileMetadata)
			filesEntrs = append(filesEntrs, fileEntry)
		case *files.FolderMetadata:
			folderEntry, _ := metadata.(*files.FolderMetadata)
			dirsEntrs = append(dirsEntrs, folderEntry)
		default:
			log.WithField("entry", metadata).Debug("unknown file entry type")
		}
	}
	return filesEntrs, dirsEntrs, nil
}

func fileToReader(token string, fileEntry *files.FileMetadata) (io.Reader, error) {
	config := dropbox.Config{
		Token:    token,
		LogLevel: dropbox.LogDebug,
	}
	dbx := files.New(config)

	downloadArgs := files.DownloadArg{
		Path: fileEntry.PathLower,
	}
	log.WithField("path", fileEntry.PathLower).Debug("downloading file")
	_, reader, err := dbx.Download(&downloadArgs)
	if err != nil {
		log.WithField("path", fileEntry.PathLower).WithError(err).Error("error downloading file")
		return reader, err
	}
	return reader, nil
}

func DownloadFile(token, localFilePath string, fileEntry *files.FileMetadata) error {
	reader, err := fileToReader(token, fileEntry)
	if err != nil {
		log.WithField("path", fileEntry.PathLower).WithError(err).Error("error downloading file")
		return err
	}

	destination, err := fs.CreateFile(localFilePath)
	if err != nil {
		log.WithError(err).WithField("path", localFilePath).Error("error creating file")
		return err
	}
	defer destination.Close()

	_, err = io.Copy(destination, reader)
	return err
}

func DownloadFileAsStr(token string, fileEntry *files.FileMetadata) (string, error) {
	reader, err := fileToReader(token, fileEntry)
	if err != nil {
		log.WithField("path", fileEntry.PathLower).WithError(err).Error("error downloading file")
		return "", err
	}

	buf := new(bytes.Buffer)
	buf.ReadFrom(reader)
	content := buf.String()
	return content, nil
}
