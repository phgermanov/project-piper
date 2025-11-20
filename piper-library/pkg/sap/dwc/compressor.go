package dwc

import (
	"archive/zip"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/SAP/jenkins-library/pkg/log"
	piperUtils "github.com/SAP/jenkins-library/pkg/piperutils"
)

type ArtifactCompressor interface {
	compressArtifact(descriptor ArtifactDescriptor, artifactSourceDir string) error
}

type fileCreator interface {
	Create(name string) (io.ReadWriteCloser, error)
}

type zipWriter interface {
	Create(name string) (io.Writer, error)
	io.Closer
}

type zipWriterFactory interface {
	NewZipWriter(writer io.Writer) zipWriter
}

type defaultZipWriterFactory struct{}

func (d defaultZipWriterFactory) NewZipWriter(writer io.Writer) zipWriter {
	return zip.NewWriter(writer)
}

type dirReader interface {
	ReadDir(dirname string) ([]fs.DirEntry, error)
}

type defaultDirReader struct{}

func (d defaultDirReader) ReadDir(dirname string) ([]fs.DirEntry, error) {
	return os.ReadDir(dirname)
}

type fileReader interface {
	ReadFile(filename string) ([]byte, error)
}

type DefaultArtifactCompressor struct {
	fileCreator
	zipWriterFactory
	dirReader
	fileReader
}

func (comp DefaultArtifactCompressor) compressArtifact(descriptor ArtifactDescriptor, artifactSourceDir string) error {
	zipFile, err := comp.Create(descriptor.getUploadFileName())
	if err != nil {
		log.SetErrorCategory(log.ErrorInfrastructure)
		return err
	}
	log.Entry().Debugf("created zip file %s successfully", descriptor.getUploadFileName())
	defer func() {
		if err := zipFile.Close(); err != nil {
			log.SetErrorCategory(log.ErrorInfrastructure)
			log.Entry().WithError(err).Error("closing zip file failed.")
		}
	}()
	zipWriter := comp.NewZipWriter(zipFile)
	defer func() {
		if err := zipWriter.Close(); err != nil {
			log.SetErrorCategory(log.ErrorInfrastructure)
			log.Entry().WithError(err).Error("closing zipfile writer failed.")
		}
	}()
	if err := comp.addFilesToZip(zipWriter, artifactSourceDir, ""); err != nil {
		log.SetErrorCategory(log.ErrorInfrastructure)
		return err
	}
	return nil
}

func (comp DefaultArtifactCompressor) addFilesToZip(zipWriter zipWriter, dir, baseInZip string) error {
	dirEnrties, err := comp.ReadDir(dir)
	if err != nil {
		return fmt.Errorf("error reading directory %s while creating zip file %w", dir, err)
	}
	for _, dirEntry := range dirEnrties {
		file, err := dirEntry.Info()
		if err != nil {
			log.Entry().Errorln("Could not read file:", file.Name())
			return err
		}

		fullFilePath := filepath.Join(dir, file.Name())
		relativeFilePath := filepath.Join(baseInZip, file.Name())

		if file.IsDir() {
			if err := comp.addFilesToZip(zipWriter, fullFilePath, relativeFilePath); err != nil {
				return err
			}
		} else {
			if err := addFileToZip(comp, zipWriter, fullFilePath, relativeFilePath); err != nil {
				return err
			}
		}
	}

	return nil
}

func addFileToZip(comp DefaultArtifactCompressor, zipWriter zipWriter, fullFilePath, relativePath string) error {
	zipFileWriter, err := zipWriter.Create(relativePath)
	if err != nil {
		return fmt.Errorf("error creating zip writer for file %s: %w", relativePath, err)
	}

	data, err := comp.ReadFile(fullFilePath)
	if err != nil {
		return fmt.Errorf("error reading file %s: %w", fullFilePath, err)
	}

	if _, err = zipFileWriter.Write(data); err != nil {
		return fmt.Errorf("error writing zip content of file %s: %w", relativePath, err)
	}

	log.Entry().Debugf("wrote file %s to zip file successfully", relativePath)

	return nil
}

func NewDefaultArtifactCompressor() DefaultArtifactCompressor {
	return DefaultArtifactCompressor{
		fileCreator:      piperUtils.Files{},
		zipWriterFactory: defaultZipWriterFactory{},
		dirReader:        defaultDirReader{},
		fileReader:       piperUtils.Files{},
	}
}
