package libutil

import (
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

func Exists(p string) (bool, error) {
	_, err := os.Stat(p)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}
	return true, err
}

func IsLink(p string) (bool, error) {
	st, err := os.Lstat(p)
	if err != nil {
		return false, err
	}
	mode := st.Mode()
	return mode&os.ModeSymlink == os.ModeSymlink, err
}

func IsEmpty(dir string) (bool, error) {
	f, err := os.Open(dir)
	if err != nil {
		return false, err
	}
	names, err := f.Readdirnames(1)
	if len(names) > 0 {
		return false, nil
	}
	if err == io.EOF {
		return true, nil
	}
	return false, err
}

func ListDir(dir string) ([]string, error) {
	var result []string
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		return result, err
	}
	for _, file := range files {
		path := filepath.Join(dir, file.Name())
		if file.Mode()&os.ModeSymlink == os.ModeSymlink {
			file, err = os.Stat(path)
			if err != nil {
				return result, err
			}
		}
		if file.IsDir() {
			if strings.HasPrefix(filepath.Base(path), ".") {
				continue
			}
			r2, err := ListDir(path)
			if err != nil {
				return result, err
			}
			result = append(result, r2...)
		} else {
			result = append(result, path)
		}
	}
	return result, nil
}
