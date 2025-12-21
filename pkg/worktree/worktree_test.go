package worktree_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/bearing-dev/bearing/pkg/worktree"
)

type testEnv struct {
	root  string
	state *worktree.State
	git   *worktree.Git
	t     *testing.T
}

func newTestEnv(t *testing.T) *testEnv {
	t.Helper()
	root := t.TempDir()

	// Create empty state file
	os.WriteFile(filepath.Join(root, "worktrees.jsonl"), nil, 0644)

	return &testEnv{
		root:  root,
		state: worktree.NewState(root),
		git:   worktree.NewGit(root),
		t:     t,
	}
}

func (e *testEnv) createRepo(name string) {
	e.t.Helper()
	repoPath := filepath.Join(e.root, name)

	run(e.t, "", "git", "init", "--initial-branch", "main", repoPath)
	os.WriteFile(filepath.Join(repoPath, "README.md"), []byte("# "+name), 0644)
	run(e.t, repoPath, "git", "add", ".")
	run(e.t, repoPath, "git", "commit", "-m", "initial")
}

func run(t *testing.T, dir string, name string, args ...string) {
	t.Helper()
	cmd := exec.Command(name, args...)
	if dir != "" {
		cmd.Dir = dir
	}
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("%s %v failed: %s: %v", name, args, out, err)
	}
}

// State tests

func TestStateReadWrite(t *testing.T) {
	env := newTestEnv(t)

	entry := worktree.Entry{
		Folder: "test-repo",
		Repo:   "test-repo",
		Branch: "main",
		Base:   true,
	}

	env.state.Append(entry)

	entries, err := env.state.Read()
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(entries))
	}
	if entries[0].Folder != "test-repo" {
		t.Errorf("expected folder test-repo, got %s", entries[0].Folder)
	}
}

func TestStateFind(t *testing.T) {
	env := newTestEnv(t)

	env.state.Append(worktree.Entry{Folder: "a", Repo: "a", Branch: "main", Base: true})
	env.state.Append(worktree.Entry{Folder: "b", Repo: "b", Branch: "main", Base: true})

	found, _ := env.state.Find("b")
	if found == nil {
		t.Fatal("expected to find entry b")
	}
	if found.Folder != "b" {
		t.Errorf("expected folder b, got %s", found.Folder)
	}

	notFound, _ := env.state.Find("c")
	if notFound != nil {
		t.Error("expected nil for non-existent entry")
	}
}

func TestStateRemove(t *testing.T) {
	env := newTestEnv(t)

	env.state.Append(worktree.Entry{Folder: "a", Repo: "a", Branch: "main", Base: true})
	env.state.Append(worktree.Entry{Folder: "b", Repo: "b", Branch: "main", Base: true})

	env.state.Remove("a")

	entries, _ := env.state.Read()
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry after removal, got %d", len(entries))
	}
	if entries[0].Folder != "b" {
		t.Errorf("expected folder b to remain, got %s", entries[0].Folder)
	}
}

// Git tests

func TestGitRepoExists(t *testing.T) {
	env := newTestEnv(t)
	env.createRepo("test-repo")

	if !env.git.RepoExists("test-repo") {
		t.Error("expected repo to exist")
	}
	if env.git.RepoExists("nonexistent") {
		t.Error("expected nonexistent repo to not exist")
	}
}

func TestGitGetBranch(t *testing.T) {
	env := newTestEnv(t)
	env.createRepo("test-repo")

	branch, err := env.git.GetBranch("test-repo")
	if err != nil {
		t.Fatal(err)
	}
	if branch != "main" {
		t.Errorf("expected branch main, got %s", branch)
	}
}

func TestGitCreateWorktree(t *testing.T) {
	env := newTestEnv(t)
	env.createRepo("test-repo")

	err := env.git.CreateWorktree("test-repo", "feature")
	if err != nil {
		t.Fatal(err)
	}

	wtPath := env.git.WorktreePath("test-repo", "feature")
	if _, err := os.Stat(wtPath); os.IsNotExist(err) {
		t.Error("worktree directory should exist")
	}

	branch, _ := env.git.GetBranch("test-repo-feature")
	if branch != "feature" {
		t.Errorf("expected branch feature, got %s", branch)
	}
}

func TestGitRemoveWorktree(t *testing.T) {
	env := newTestEnv(t)
	env.createRepo("test-repo")
	env.git.CreateWorktree("test-repo", "feature")

	err := env.git.RemoveWorktree("test-repo", "feature")
	if err != nil {
		t.Fatal(err)
	}

	wtPath := env.git.WorktreePath("test-repo", "feature")
	if _, err := os.Stat(wtPath); !os.IsNotExist(err) {
		t.Error("worktree directory should not exist after removal")
	}
}

func TestIsMainBranch(t *testing.T) {
	if !worktree.IsMainBranch("main") {
		t.Error("main should be main branch")
	}
	if !worktree.IsMainBranch("master") {
		t.Error("master should be main branch")
	}
	if worktree.IsMainBranch("feature") {
		t.Error("feature should not be main branch")
	}
}

func TestGitIsDirty(t *testing.T) {
	env := newTestEnv(t)
	env.createRepo("test-repo")

	// Clean repo should not be dirty
	dirty, err := env.git.IsDirty("test-repo")
	if err != nil {
		t.Fatal(err)
	}
	if dirty {
		t.Error("clean repo should not be dirty")
	}

	// Add uncommitted file
	repoPath := filepath.Join(env.root, "test-repo")
	os.WriteFile(filepath.Join(repoPath, "new-file.txt"), []byte("content"), 0644)

	dirty, err = env.git.IsDirty("test-repo")
	if err != nil {
		t.Fatal(err)
	}
	if !dirty {
		t.Error("repo with uncommitted file should be dirty")
	}
}
