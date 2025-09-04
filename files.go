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
func (h *GoBuild) renameOutputFile(tempFileName string) error {
	tempPath := path.Join(h.config.OutFolderRelativePath, tempFileName)
	finalPath := path.Join(h.config.OutFolderRelativePath, h.outFileName)

	// fmt.Fprintf(h.config.Logger, "Renaming %s to %s\n", tempPath, finalPath)

	err := os.Rename(tempPath, finalPath)
	if err != nil {
		if h.config.Logger != nil {
			h.config.Logger("Rename failed:", err)
		}
		return errors.Join(errors.New("renameOutputFile"), err)
	}

	// fmt.Fprintf(h.config.Logger, "Rename successful\n")

	return nil
}

// cleanupTempFile removes the temporary output file if it exists
// This is called when compilation fails to ensure no partial files remain
func (h *GoBuild) cleanupTempFile(tempFileName string) {
	tempFilePath := path.Join(h.config.OutFolderRelativePath, tempFileName)
	if _, err := os.Stat(tempFilePath); err == nil {
		// File exists, try to remove it
		os.Remove(tempFilePath)
		// We don't handle the error here as it's a cleanup operation
		// and the main error (compilation failure) is more important
	}
}
