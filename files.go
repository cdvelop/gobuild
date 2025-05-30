package gobuild

import (
	"errors"
	"os"
	"path"
)

// UnobservedFiles returns the list of files that should not be tracked by file watchers
// eg: main.exe, main_temp.exe
func (h *GoBuild) UnobservedFiles() []string {
	return []string{
		h.outFileName,
		h.outTempFileName,
	}
}

// renameOutputFile renames the temporary output file to the final output file
func (h *GoBuild) renameOutputFile() error {
	err := os.Rename(
		path.Join(h.OutFolder, h.outTempFileName),
		path.Join(h.OutFolder, h.outFileName),
	)
	if err != nil {
		return errors.Join(errors.New("renameOutputFile"), err)
	}
	return nil
}

// cleanupTempFile removes the temporary output file if it exists
// This is called when compilation fails to ensure no partial files remain
func (h *GoBuild) cleanupTempFile() {
	tempFilePath := path.Join(h.OutFolder, h.outTempFileName)
	if _, err := os.Stat(tempFilePath); err == nil {
		// File exists, try to remove it
		os.Remove(tempFilePath)
		// We don't handle the error here as it's a cleanup operation
		// and the main error (compilation failure) is more important
	}
}
