//go:build unit
// +build unit

package dwc

import (
	"bytes"
	"io/fs"
	"path/filepath"
	"testing"

	piperMocks "github.com/SAP/jenkins-library/pkg/mock"
)

func TestDefaultArtifactCompressor_compressArtifact(t *testing.T) {
	t.Parallel()
	type fsMeta struct {
		fileContents   []byte
		assignedWriter *zipFileWriterMock
	}
	type testCase struct {
		name                string
		artifactSourceDir   string
		zipFile             *readerWriterCloserMock
		configureDescriptor func() ArtifactDescriptor
		configureCompressor func(self *testCase, t *testing.T) DefaultArtifactCompressor
		verify              func(self *testCase, t *testing.T)
		fileContents        map[string]fsMeta
		wantErr             bool
	}
	tests := []*testCase{
		{
			name:              "canonical compress",
			artifactSourceDir: "/dev/null",
			configureDescriptor: func() ArtifactDescriptor {
				descriptor := &artifactDescriptorMock{}
				descriptor.On("getUploadFileName").Return("upload.zip")
				return descriptor
			},
			configureCompressor: func(self *testCase, t *testing.T) DefaultArtifactCompressor {
				t.Helper()
				// uploadFile should be created as defined by the descriptor
				self.zipFile = &readerWriterCloserMock{}
				fileCreator := newMockFileCreator(t)
				fileCreator.On("Create", "upload.zip").Return(self.zipFile, nil)
				// a new zip writer should be created based on the created file
				// the zip writer should create the files that are children of artifactSourceDir
				zipWriter := newMockZipWriter(t)
				zipWriter.On("Create", "directChildrenFile.txt").Return(self.fileContents["directChildrenFile.txt"].assignedWriter, nil)
				zipWriter.On("Create", "subFolder/nestedChildrenFile.txt").Return(self.fileContents["subFolder/nestedChildrenFile.txt"].assignedWriter, nil)
				zipWriterFactory := newMockZipWriterFactory(t)
				zipWriterFactory.On("NewZipWriter", self.zipFile).Return(zipWriter)
				// the zip writer should be closed
				zipWriter.On("Close").Return(nil)
				// the directory artifactSourceDir should be read from disk to compress the files beneath it
				dirReader := newMockDirReader(t)
				dirReader.On("ReadDir", self.artifactSourceDir).Return([]fs.DirEntry{
					&DirEntryMock{
						fileInfo: &fileInfoMock{
							fileName: "directChildrenFile.txt",
							isDir:    false,
						},
					},
					&DirEntryMock{
						fileInfo: &fileInfoMock{
							fileName: "subFolder",
							isDir:    true,
						},
					},
				}, nil)
				// the directory subFolder should be read from disk to compress the files beneath it
				dirReader.On("ReadDir", filepath.Join(self.artifactSourceDir, "subFolder")).Return([]fs.DirEntry{
					&DirEntryMock{
						fileInfo: &fileInfoMock{
							fileName: "nestedChildrenFile.txt",
							isDir:    false,
						},
					},
				}, nil)
				// every file of the dir artifactSourceDir should actually be written to the zip files contents
				filesMock := &piperMocks.FilesMock{}
				filesMock.AddFile("/dev/null/directChildrenFile.txt", self.fileContents["directChildrenFile.txt"].fileContents)
				filesMock.AddFile("/dev/null/subFolder/nestedChildrenFile.txt", self.fileContents["subFolder/nestedChildrenFile.txt"].fileContents)
				return DefaultArtifactCompressor{
					fileCreator:      fileCreator,
					zipWriterFactory: zipWriterFactory,
					dirReader:        dirReader,
					fileReader:       filesMock,
				}
			},
			verify: func(self *testCase, t *testing.T) {
				t.Helper()
				// the zip file should be closed after writing its contents to the fs
				if !self.zipFile.closed {
					t.Errorf("Expected zip file to be closed but it never was")
				}
				// zip data written should match the bytes of the files consumed
				if !bytes.Equal(self.fileContents["directChildrenFile.txt"].assignedWriter.buffer.Bytes(), self.fileContents["directChildrenFile.txt"].fileContents) {
					t.Errorf("data written to zip for file %s is %s, but expected was %s", "directChildrenFile.txt", self.fileContents["directChildrenFile.txt"].assignedWriter.buffer.Bytes(), self.fileContents["directChildrenFile.txt"].fileContents)
				}
				if !bytes.Equal(self.fileContents["subFolder/nestedChildrenFile.txt"].assignedWriter.buffer.Bytes(), self.fileContents["subFolder/nestedChildrenFile.txt"].fileContents) {
					t.Errorf("data written to zip for file %s is %s, but expected was %s", "subFolder/nestedChildrenFile.txt", self.fileContents["subFolder/nestedChildrenFile.txt"].assignedWriter.buffer.Bytes(), self.fileContents["subFolder/nestedChildrenFile.txt"].fileContents)
				}
			},
			fileContents: map[string]fsMeta{
				"directChildrenFile.txt": {
					fileContents: []byte("Some data for the direct child"),
					assignedWriter: &zipFileWriterMock{
						buffer: &bytes.Buffer{},
					},
				},
				"subFolder/nestedChildrenFile.txt": {
					fileContents: []byte("Some data for the nested child"),
					assignedWriter: &zipFileWriterMock{
						buffer: &bytes.Buffer{},
					},
				},
			},
			wantErr: false,
		},
	}
	for _, c := range tests {
		testCase := c
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			compressor := testCase.configureCompressor(testCase, t)
			err := compressor.compressArtifact(testCase.configureDescriptor(), testCase.artifactSourceDir)
			if (err != nil) != testCase.wantErr {
				t.Fatalf("compressArtifact() error = %v, wantErr %v", err, testCase.wantErr)
			}
			testCase.verify(testCase, t)
		})
	}
}
