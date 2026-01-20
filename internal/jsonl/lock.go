package jsonl

import (
	"os"
	"path/filepath"
	"syscall"
)

// FileLock provides flock-based file locking
type FileLock struct {
	file *os.File
}

// NewFileLock creates a lock file at the given path
func NewFileLock(path string) (*FileLock, error) {
	// Ensure parent directory exists
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, err
	}

	f, err := os.OpenFile(path+".lock", os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		return nil, err
	}
	return &FileLock{file: f}, nil
}

// Lock acquires an exclusive lock
func (l *FileLock) Lock() error {
	return syscall.Flock(int(l.file.Fd()), syscall.LOCK_EX)
}

// RLock acquires a shared (read) lock
func (l *FileLock) RLock() error {
	return syscall.Flock(int(l.file.Fd()), syscall.LOCK_SH)
}

// Unlock releases the lock
func (l *FileLock) Unlock() error {
	return syscall.Flock(int(l.file.Fd()), syscall.LOCK_UN)
}

// Close releases the lock and closes the file
func (l *FileLock) Close() error {
	l.Unlock()
	return l.file.Close()
}
