/**
* Copyright 2019 gd Author. All rights reserved.
* Author: Chuck1024
 */

package dlog

import (
	"fmt"
	"log"
	"os"
	"strings"
	"sync"
	"time"
)

// This log writer sends output to a file
type FileLogWriter struct {
	rec  chan *LogRecord
	stop chan bool
	rot  chan bool

	// The opened file
	filename      string
	file          *os.File
	fileCloseLock sync.Mutex
	closeOnce     sync.Once

	// The logging format
	format string

	// File header/trailer
	header, trailer string

	// Rotate at lineCount
	maxLines         int
	maxLinesCurLines int

	// Rotate at size
	maxsize        int
	maxsizeCurSize int

	// Rotate daily
	daily         bool
	dailyOpenDate int

	// rotate hourly
	hourly         bool
	hourlyOpenHour int

	// Keep old logfiles (.001, .002, etc)
	rotate         bool
	ScribeCategory string

	formatCache formatCacheType
}

// This is the FileLogWriter's output method
func (w *FileLogWriter) LogWrite(rec *LogRecord) {
	select {
	case w.rec <- rec:
	case <-w.stop:
		log.Println(fmt.Sprintf("write on closed logger:%v", rec))
	default:
		select {
		case w.rec <- rec:
		case <-w.stop:
			log.Println(fmt.Sprintf("write on closed logger:%v", rec))
		case <-time.After(2 * time.Millisecond):
			//add "fatal " prefix in stderr log to trigger sms alert
			fmt.Fprintf(os.Stderr, "fatal file log channel blocked!%v\n", rec)
		}
	}
}

func (w *FileLogWriter) Close() {
	w.closeOnce.Do(func() {
		w.fileCloseLock.Lock()
		defer w.fileCloseLock.Unlock()
		close(w.stop)
		w.file.Sync()
	})
}

// NewFileLogWriter creates a new LogWriter which writes to the given file and
// has rotation enabled if rotate is true.
//
// If rotate is true, any time a new log file is opened, the old one is renamed
// with a .### extension to preserve it.  The various Set* methods can be used
// to configure log rotation based on lines, size, and daily.
//
// The standard log-line format is:
//   [%D %T] [%L] (%S) %M
func NewFileLogWriter(fileName string, rotate, daily, hourly bool) *FileLogWriter {
	w := &FileLogWriter{
		rec:         make(chan *LogRecord, LogBufferLength),
		stop:        make(chan bool),
		rot:         make(chan bool),
		filename:    fileName,
		format:      "[%D %T] [%L] (%S) %M",
		rotate:      rotate,
		daily:       daily,
		hourly:      hourly,
		formatCache: formatCacheType{},
	}

	// open the file for the first time
	if err := w.intRotate(); err != nil {
		fmt.Fprintf(os.Stderr, "FileLogWriter(%q): %s\n", w.filename, err)
		return nil
	}

	go func() {
		defer func() {
			if w.file != nil {
				fmt.Fprint(w.file, FormatLogRecord(&w.formatCache, w.trailer, &LogRecord{Created: time.Now()}))
				w.fileCloseLock.Lock()
				w.file.Close()
				w.fileCloseLock.Unlock()
			}
		}()

		for {
			select {
			case <-w.stop:
				return
			case <-w.rot:
				if err := w.intRotate(); err != nil {
					fmt.Fprintf(os.Stderr, "FileLogWriter(%q): %s\n", w.filename, err)
					// BUG FIX: if this err happens, panic.
					//          just print err msg and return will cause all goroutine hanged in printing log.
					panic(fmt.Sprintf("FileLogWriter(%q): %s\n", w.filename, err))
				}
			case rec, ok := <-w.rec:
				if !ok {
					return
				}
				now := &rec.Created
				if (w.maxLines > 0 && w.maxLinesCurLines >= w.maxLines) ||
					(w.maxsize > 0 && w.maxsizeCurSize >= w.maxsize) ||
					(w.daily && now.Day() != w.dailyOpenDate) ||
					(w.hourly && now.Hour() != w.hourlyOpenHour) {
					if err := w.intRotateTime(now); err != nil {
						fmt.Fprintf(os.Stderr, "FileLogWriter(%q): %s\n", w.filename, err)
						// BUG FIX: if this err happens, panic
						panic(fmt.Sprintf("FileLogWriter(%q): %s\n", w.filename, err))
					}
				}

				// Perform the write
				toWrite := FormatLogRecord(&w.formatCache, w.format, rec)
				n, err := fmt.Fprint(w.file, toWrite)
				if err != nil {
					fmt.Fprintf(os.Stderr, "FileLogWriter(%q): %s\n", w.filename, err)
					// BUG FIX: if this err happens, panic
					panic(fmt.Sprintf("FileLogWriter(%q): %s\n", w.filename, err))
				}

				// Update the counts
				w.maxLinesCurLines++
				w.maxsizeCurSize += n
			}
		}
	}()

	return w
}

// Request that the logs rotate
func (w *FileLogWriter) Rotate() {
	select {
	case w.rot <- true:
	case <-w.stop:
	}
}

// If this is called in a threaded context, it MUST be synchronized
func (w *FileLogWriter) intRotate() error {
	now := time.Now()
	return w.intRotateTime(&now)
}

func (w *FileLogWriter) intRotateTime(now *time.Time) error {
	// Close any log file that may be open
	if w.file != nil {
		fmt.Fprint(w.file, FormatLogRecord(&w.formatCache, w.trailer, &LogRecord{Created: time.Now()}))
		w.fileCloseLock.Lock()
		w.file.Close()
		w.fileCloseLock.Unlock()
	}

	if now == nil {
		_now := time.Now()
		now = &_now
	}
	// If we are keeping log files, move it to the next available number
	if w.rotate {
		_, err := os.Lstat(w.filename)
		if err == nil { // file exists
			// Find the next available number
			num := 0
			fName := ""
			if w.daily && now.Day() != w.dailyOpenDate {
				yesterday := now.AddDate(0, 0, -1).Format("20060102")
				for ; err == nil && num <= 999; num++ {
					if num == 0 {
						fName = w.filename + fmt.Sprintf(".%s", yesterday)
					} else {
						fName = w.filename + fmt.Sprintf(".%s.%03d", yesterday, num)
					}
					_, err = os.Lstat(fName)
				}
			} else if w.hourly && now.Hour() != w.hourlyOpenHour {
				lastHour := now.Add(-1 * time.Hour).Format("2006010215")
				for ; err == nil && num <= 999; num++ {
					if num == 0 {
						fName = w.filename + fmt.Sprintf(".%s", lastHour)
					} else {
						fName = w.filename + fmt.Sprintf(".%s.%03d", lastHour, num)
					}
					_, err = os.Lstat(fName)
				}
			} else {
				for ; err == nil && num <= 999; num++ {
					fName = w.filename + fmt.Sprintf(".%s.%03d", now.Format("2006010215"), num)
					_, err = os.Lstat(fName)
				}
			}
			// return error if the last file checked still existed
			if err == nil {
				return fmt.Errorf("Rotate: Cannot find free log number to rename %s\n", w.filename)
			}
			w.fileCloseLock.Lock()
			w.file.Close()
			w.fileCloseLock.Unlock()
			// Rename the file to its newfound home
			err = os.Rename(w.filename, fName)
			if err != nil {
				return fmt.Errorf("Rotate: %s\n", err)
			}
		}
	}

	// Open the log file
	ss := strings.Split(w.filename, "/")
	ss = ss[:len(ss)-1]
	if len(ss) >= 1 {
		s := strings.Join(ss, "/")
		if _, err := os.Stat(s); err != nil {
			err := os.MkdirAll(s, 0777)
			if err != nil {
				log.Printf("Error creating directory, err:%s", err)
				return err
			}
			_, err = os.Create(w.filename)
			if err != nil {
				log.Printf("Error creating file, err:%s", err)
				return err
			}
		}
	}

	if _, err := os.Stat(w.filename); err != nil {
		_, err = os.Create(w.filename)
		if err != nil {
			log.Printf("Error creating file, err:%s", err)
			return err
		}
	}

	fd, err := os.OpenFile(w.filename, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0666)
	if err != nil {
		return err
	}
	w.fileCloseLock.Lock()
	w.file = fd
	w.fileCloseLock.Unlock()

	fmt.Fprint(w.file, FormatLogRecord(&w.formatCache, w.header, &LogRecord{Created: *now}))

	// Set the daily open date to the current date
	w.dailyOpenDate = now.Day()
	w.hourlyOpenHour = now.Hour()

	// initialize rotation values
	w.maxLinesCurLines = 0
	w.maxsizeCurSize = 0

	return nil
}

// Set the logging format (chainable).  Must be called before the first log
// message is written.
func (w *FileLogWriter) SetFormat(format string) *FileLogWriter {
	w.format = format
	return w
}

// Set the logfile header and footer (chainable).  Must be called before the first log
// message is written.  These are formatted similar to the FormatLogRecord (e.g.
// you can use %D and %T in your header/footer for date and time).
func (w *FileLogWriter) SetHeadFoot(head, foot string) *FileLogWriter {
	w.header, w.trailer = head, foot
	if w.maxLinesCurLines == 0 {
		fmt.Fprint(w.file, FormatLogRecord(&w.formatCache, w.header, &LogRecord{Created: time.Now()}))
	}
	return w
}

// Set rotate at linecount (chainable). Must be called before the first log
// message is written.
func (w *FileLogWriter) SetRotateLines(maxLines int) *FileLogWriter {
	//fmt.Fprintf(runtime.Stderr, "FileLogWriter.SetRotateLines: %v\n", maxLines)
	w.maxLines = maxLines
	return w
}

// Set rotate at size (chainable). Must be called before the first log message
// is written.
func (w *FileLogWriter) SetRotateSize(maxsize int) *FileLogWriter {
	//fmt.Fprintf(runtime.Stderr, "FileLogWriter.SetRotateSize: %v\n", maxsize)
	w.maxsize = maxsize
	return w
}

// Set rotate daily (chainable). Must be called before the first log message is
// written.
func (w *FileLogWriter) SetRotateDaily(daily bool) *FileLogWriter {
	//fmt.Fprintf(runtime.Stderr, "FileLogWriter.SetRotateDaily: %v\n", daily)
	w.daily = daily
	return w
}

func (w *FileLogWriter) SetRotateHourly(hourly bool) *FileLogWriter {
	//fmt.Fprintf(runtime.Stderr, "FileLogWriter.SetRotateHourly: %v\n", hourly)
	w.hourly = hourly
	return w
}

// SetRotate changes whether or not the old logs are kept. (chainable) Must be
// called before the first log message is written.  If rotate is false, the
// files are overwritten; otherwise, they are rotated to another file before the
// new log is opened.
func (w *FileLogWriter) SetRotate(rotate bool) *FileLogWriter {
	//fmt.Fprintf(runtime.Stderr, "FileLogWriter.SetRotate: %v\n", rotate)
	w.rotate = rotate
	return w
}

// NewXMLLogWriter is a utility method for creating a FileLogWriter set up to
// output XML record log messages instead of line-based ones.
func NewXMLLogWriter(fName string, rotate bool) *FileLogWriter {
	return NewFileLogWriter(fName, rotate, false, true).SetFormat(
		`	<record level="%L">
		<timestamp>%D %T</timestamp>
		<source>%S</source>
		<message>%M</message>
	</record>`).SetHeadFoot("<log created=\"%D %T\">", "</log>")
}
