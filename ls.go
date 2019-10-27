package script

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
)

// Files is a stream of a list of files. A user can eigher use the file list directly or the the
// created stream. In the stream, each line contains a path to a file.
type Files struct {
	Stream
	Files []File
}

// File contains information about a file.
type File struct {
	// FileInfo contains information about the file.
	os.FileInfo
	// Path is the path of the file. It may be relative or absolute, depending on how the `Ls`
	// command was invoked.
	Path string
}

// Ls returns a stream of a list files. In the returned stream, each line will contain a path to
// a single file.
//
// If the provided paths list is empty, the local directory will be listed.
//
// The provided paths may be relative to the local directory or absolute - this will influence the
// format of the returned paths in the output.
//
// If some provided paths correlate to the arguments correlate to the same file, it will also appear
// multiple times in the output.
//
// If any of the paths fails to be listed, it will result in an error in the output, but the stream
// will still conain all paths that were successfully listed.
//
// Shell command: `ls`.
func Ls(paths ...string) Files {
	// Default to local directory.
	if len(paths) == 0 {
		paths = append(paths, ".")
	}

	var (
		command = Command{Name: fmt.Sprintf("ls (%+v)", paths)}
		files   []File
	)

	for _, path := range paths {
		info, err := os.Stat(path)
		if err != nil {
			command.AppendError(err, "stat path")
			continue
		}

		// Path is a single file.
		if !info.IsDir() {
			files = append(files, File{Path: path, FileInfo: info})
			continue
		}

		// Path is a directory.
		infos, err := ioutil.ReadDir(path)
		if err != nil {
			command.AppendError(err, "read dir")
			continue
		}

		for _, info := range infos {
			files = append(files, File{Path: filepath.Join(path, info.Name()), FileInfo: info})
		}
	}
	command.Reader = &filesReader{files: files}

	return Files{
		Stream: Stdin().PipeTo(func(io.Reader) Command { return command }),
		Files:  files,
	}
}

// filesReader reads from a file info list.
type filesReader struct {
	files []File
	// seek indicates which file to write for the next Read function call.
	seek int
}

func (f *filesReader) Read(out []byte) (int, error) {
	if f.seek >= len(f.files) {
		return 0, io.EOF
	}

	line := []byte(f.files[f.seek].Path + "\n")
	f.seek++

	n := copy(out, line)
	return n, nil
}
