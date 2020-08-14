package helpers

import (
	"github.com/minio/minio-go/v6"
	"io"
	"net/url"
	"time"
)

type FileStorage struct {
	client *minio.Client
}

func (f *FileStorage) CreateFolder(folderName string) error {
	err := f.client.MakeBucket(folderName, "")
	if err != nil {
		// Check to see if we already own this bucket (which happens if you run this twice)
		exists, errBucketExists := f.client.BucketExists(folderName)
		if errBucketExists == nil && exists {
			return nil
		} else {
			return err
		}
	}
	return nil
}

func (f *FileStorage) UploadFile(folderName string, fileName string, file io.Reader, size int64) error {
	_, err := f.client.PutObject(folderName,
		fileName,
		file,
		size,
		minio.PutObjectOptions{ContentType: "application/octet-stream"},
	)
	if err != nil {
		return err
	}
	return nil
}

func (f *FileStorage) DownloadFile(folderName string, fileName string, saveFolder string) error {
	err := f.client.FGetObject(folderName, fileName, saveFolder, minio.GetObjectOptions{})
	if err != nil {
		return err
	}
	return nil
}

func (f *FileStorage) GetFileLink(folderName string, filename string, expires time.Duration) (string, error) {
	reqParams := make(url.Values)
	url, err := f.client.PresignedGetObject(folderName,
		filename,
		expires,
		reqParams,
	)
	if err != nil {
		return "", err
	}
	return url.Host + url.Path + "?" + url.RawQuery, nil
}

func (f *FileStorage) RemoveFile(folderName string, filename string) error {
	err := f.client.RemoveObject(folderName, filename)
	return err
}

func (f *FileStorage) RemoveBucket(bucketName string) error {
	err := f.client.RemoveBucket(bucketName)
	return err
}
