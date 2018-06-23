package logging

import (
	"io"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

type fileNameInfo struct {
	fileName string
	fileInfo os.FileInfo
}

type fileNameInfoSlice struct {
	files []fileNameInfo
}

func (f *fileNameInfoSlice) Len() int {
	return len(f.files)
}

func (f *fileNameInfoSlice) Less(i, j int) bool {
	return f.files[i].fileInfo.ModTime().Before(f.files[j].fileInfo.ModTime())
}

func (f *fileNameInfoSlice) Swap(i, j int) {
	f.files[j], f.files[i] = f.files[i], f.files[j]
}

func (f *fileNameInfoSlice) Sort() {
	sort.Sort(f)
}

func (f *fileNameInfoSlice) removeBefore(n int) {
	files := []fileNameInfo{}
	for i := 0; i < f.Len(); i++ {
		if i < n {
			_ = os.Remove(f.files[i].fileName)
		} else {
			files = append(files, f.files[i])
		}
	}
	f.files = files
}

func (f *fileNameInfoSlice) renameIndex(prefix string) {
	for index, fi := range f.files {
		newname := prefix + "." + strconv.Itoa(index+1)
		_ = os.Rename(fi.fileName, newname)
	}
}

type SizeRotationHandler struct {
	*Handler
	fileName    string
	curFileSize uint64
	maxFileSize uint64
	maxFiles    uint32
}

func NewSizeRotationHandler(fn string, size uint64, count uint32) (*SizeRotationHandler, error) {
	h := &SizeRotationHandler{fileName: fn, maxFileSize: size, maxFiles: count}
	fp, err := h.openCreateFile(fn)
	if err != nil {
		return nil, err
	}
	h.curFileSize, err = h.fileSize()
	if err != nil {
		_ = fp.Close()
		return nil, err
	}
	h.Handler = NewHandler(fp)
	h.before = h.rotate
	h.after = h.afterWrite
	return h, nil
}

func (h *SizeRotationHandler) openCreateFile(fn string) (*os.File, error) {
	return os.OpenFile(fn, FileCreateFlag, FileCreatePerm)
}

func (h *SizeRotationHandler) fileSize() (uint64, error) {
	info, err := os.Stat(h.fileName)
	if err != nil {
		return 0, err
	}
	return uint64(info.Size()), nil
}

func (h *SizeRotationHandler) releaseFiles() (string, error) {
	pattern := h.fileName + ".*"
	fs, err := filepath.Glob(pattern)
	if err != nil {
		return "", err
	}
	re, err2 := regexp.Compile("[0-9]+")
	if err2 != nil {
		return "", err2
	}
	files := &fileNameInfoSlice{}
	for _, name := range fs {
		suf := strings.TrimPrefix(name+".", h.fileName)
		if re.MatchString(suf) {
			if fileinfo, err := os.Stat(name); err == nil {
				files.files = append(files.files, fileNameInfo{name, fileinfo})
			}
		}
	}
	files.Sort()
	files.removeBefore(files.Len() - int(h.maxFiles))
	files.renameIndex(h.fileName)
	release := h.fileName + "." + strconv.Itoa(files.Len()+1)
	return release, nil
}

func (h *SizeRotationHandler) afterWrite(rd *Record, n int64) {
	h.curFileSize += uint64(n)
}

func (h *SizeRotationHandler) rotate(*Record, io.ReadWriter) {
	if h.curFileSize < h.maxFileSize {
		return
	}
	h.curFileSize = 0
	_ = h.writer.(io.Closer).Close()
	name, err := h.releaseFiles()
	if err != nil {
		return
	}
	if err := os.Rename(h.fileName, name); err != nil {
		return
	}
	fp, err := h.openCreateFile(h.fileName)
	if err != nil {
		return
	}
	h.writer = fp
}
