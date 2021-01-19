package zipdownload

import (
	"archive/zip"
	"bytes"
	"io/ioutil"

	"github.com/zachmann/mytoken/internal/httpClient"
)

// DownloadZipped downloads a zip archive and returns all contained files
func DownloadZipped(url string) (map[string][]byte, error) {
	resp, err := httpClient.Do().R().Get(url)
	if err != nil {
		return nil, err
	}

	body := resp.Body()
	zipReader, err := zip.NewReader(bytes.NewReader(body), int64(len(body)))
	if err != nil {
		return nil, err
	}

	allFiles := make(map[string][]byte)
	// Read all the files from zip archive
	for _, zipFile := range zipReader.File {
		unzippedFileBytes, err := readZipFile(zipFile)
		if err != nil {
			return allFiles, err
		}
		allFiles[zipFile.Name] = unzippedFileBytes
	}
	return allFiles, nil
}

func readZipFile(zf *zip.File) ([]byte, error) {
	f, err := zf.Open()
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return ioutil.ReadAll(f)
}
