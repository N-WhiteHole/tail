// TODO: avoid creating two instances of the fsnotify.Watcher struct
package tail

import (
	"github.com/howeyc/fsnotify"
	"os"
	"path/filepath"
	"time"
)

type FileWatcher interface {
	BlockUntilExists() error
	ChangeEvents() chan bool
}

// FileWatcher monitors file-level events
type InotifyFileWatcher struct {
	Filename string
}

func NewInotifyFileWatcher(filename string) *InotifyFileWatcher {
	fw := &InotifyFileWatcher{filename}
	return fw
}

// BlockUntilExists blocks until the file comes into existence. If the
// file already exists, then block until it is created again.
func (fw *InotifyFileWatcher) BlockUntilExists() error {
	w, err := fsnotify.NewWatcher()
	if err != nil {
		return err
	}
	defer w.Close()
	err = w.WatchFlags(filepath.Dir(fw.Filename), fsnotify.FSN_CREATE)
	if err != nil {
		return err
	}
	defer w.RemoveWatch(filepath.Dir(fw.Filename))
	<-w.Event
	return nil
}

// ChangeEvents returns a channel that gets updated when the file is ready to be read.
func (fw *InotifyFileWatcher) ChangeEvents() chan bool {
	w, err := fsnotify.NewWatcher()
	if err != nil {
		panic(err)
	}
	err = w.Watch(fw.Filename)
	if err != nil {
		panic(err)
	}

	ch := make(chan bool)

	go func() {
		for {
			evt := <-w.Event
			switch {
			case evt.IsDelete():
				fallthrough

			case evt.IsRename():
				close(ch)
				w.RemoveWatch(fw.Filename)
				w.Close()
				return

			case evt.IsModify():
				// send only if channel is empty.
				select {
				case ch <- true:
				default:
				}
			}
		}
	}()

	return ch
}

// FileWatcher monitors file-level events
type PollingFileWatcher struct {
	Filename string
}

func NewPollingFileWatcher(filename string) *PollingFileWatcher {
	fw := &PollingFileWatcher{filename}
	return fw
}

// BlockUntilExists blocks until the file comes into existence. If the
// file already exists, then block until it is created again.
func (fw *PollingFileWatcher) BlockUntilExists() error {
	panic("not implemented")
	return nil
}

// ChangeEvents returns a channel that gets updated when the file is ready to be read.
func (fw *PollingFileWatcher) ChangeEvents() chan bool {
	ch := make(chan bool)
	stop := make(chan bool)
	every2Seconds := time.Tick(2 * time.Second)

	go func() {
		for {
			time.Sleep(250 * time.Millisecond)
			select {
			case ch <- true:
			case <-stop:
				return
			default:
			}
		}
	}()

	go func() {
		for {
			select {
			case <-every2Seconds:
				// NOTE: not using file descriptor.
				if _, err := os.Stat(fw.Filename); os.IsNotExist(err) {
					stop <- true
					close(ch)
					return
				}
			}
		}
	}()

	return ch
}