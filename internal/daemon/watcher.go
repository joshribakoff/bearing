package daemon

import (
	"path/filepath"

	"github.com/fsnotify/fsnotify"
)

// Watcher wraps fsnotify to watch for file changes
type Watcher struct {
	watcher   *fsnotify.Watcher
	onChange  func()
	watchPath string
}

// NewWatcher creates a new file watcher for the workspace
func NewWatcher(workspaceDir string, onChange func()) (*Watcher, error) {
	w, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}

	watcher := &Watcher{
		watcher:   w,
		onChange:  onChange,
		watchPath: workspaceDir,
	}

	// Watch local.jsonl for changes
	localPath := filepath.Join(workspaceDir, "local.jsonl")
	if err := w.Add(localPath); err != nil {
		// File might not exist yet, that's ok
	}

	return watcher, nil
}

// Start begins watching for changes
func (w *Watcher) Start() {
	go func() {
		for {
			select {
			case event, ok := <-w.watcher.Events:
				if !ok {
					return
				}
				if event.Op&fsnotify.Write == fsnotify.Write {
					w.onChange()
				}
			case _, ok := <-w.watcher.Errors:
				if !ok {
					return
				}
			}
		}
	}()
}

// Close stops watching
func (w *Watcher) Close() error {
	return w.watcher.Close()
}
