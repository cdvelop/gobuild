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
