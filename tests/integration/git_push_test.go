// Copyright 2024 The Gitea Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package integration

import (
	"fmt"
	"net/url"
	"testing"

	"code.gitea.io/gitea/models/db"
	git_model "code.gitea.io/gitea/models/git"
	"code.gitea.io/gitea/models/unittest"
	user_model "code.gitea.io/gitea/models/user"
	"code.gitea.io/gitea/modules/git"
	repo_service "code.gitea.io/gitea/services/repository"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGitPush(t *testing.T) {
	onGiteaRun(t, testGitPush)
}

func testGitPush(t *testing.T, u *url.URL) {
	t.Run("Push branches at once", func(t *testing.T) {
		runTestGitPush(t, u, func(t *testing.T, gitPath string) (pushed, deleted []string) {
			for i := 0; i < 100; i++ {
				branchName := fmt.Sprintf("branch-%d", i)
				pushed = append(pushed, branchName)
				doGitCreateBranch(gitPath, branchName)(t)
			}
			pushed = append(pushed, "master")
			doGitPushTestRepository(gitPath, "origin", "--all")(t)
			return pushed, deleted
		})
	})

	t.Run("Push branches one by one", func(t *testing.T) {
		runTestGitPush(t, u, func(t *testing.T, gitPath string) (pushed, deleted []string) {
			for i := 0; i < 100; i++ {
				branchName := fmt.Sprintf("branch-%d", i)
				doGitCreateBranch(gitPath, branchName)(t)
				doGitPushTestRepository(gitPath, "origin", branchName)(t)
				pushed = append(pushed, branchName)
			}
			return pushed, deleted
		})
	})

	t.Run("Delete branches", func(t *testing.T) {
		runTestGitPush(t, u, func(t *testing.T, gitPath string) (pushed, deleted []string) {
			doGitPushTestRepository(gitPath, "origin", "master")(t) // make sure master is the default branch instead of a branch we are going to delete
			pushed = append(pushed, "master")

			for i := 0; i < 100; i++ {
				branchName := fmt.Sprintf("branch-%d", i)
				pushed = append(pushed, branchName)
				doGitCreateBranch(gitPath, branchName)(t)
			}
			doGitPushTestRepository(gitPath, "origin", "--all")(t)

			for i := 0; i < 10; i++ {
				branchName := fmt.Sprintf("branch-%d", i)
				doGitPushTestRepository(gitPath, "origin", "--delete", branchName)(t)
				deleted = append(deleted, branchName)
			}
			return pushed, deleted
		})
	})
}

func runTestGitPush(t *testing.T, u *url.URL, gitOperation func(t *testing.T, gitPath string) (pushed, deleted []string)) {
	user := unittest.AssertExistsAndLoadBean(t, &user_model.User{ID: 2})
	repo, err := repo_service.CreateRepository(db.DefaultContext, user, user, repo_service.CreateRepoOptions{
		Name:          "repo-to-push",
		Description:   "test git push",
		AutoInit:      false,
		DefaultBranch: "main",
		IsPrivate:     false,
	})
	require.NoError(t, err)
	require.NotEmpty(t, repo)

	gitPath := t.TempDir()

	doGitInitTestRepository(gitPath)(t)

	oldPath := u.Path
	oldUser := u.User
	defer func() {
		u.Path = oldPath
		u.User = oldUser
	}()
	u.Path = repo.FullName() + ".git"
	u.User = url.UserPassword(user.LowerName, userPassword)

	doGitAddRemote(gitPath, "origin", u)(t)

	gitRepo, err := git.OpenRepository(git.DefaultContext, gitPath)
	require.NoError(t, err)
	defer gitRepo.Close()

	pushedBranches, deletedBranches := gitOperation(t, gitPath)

	dbBranches := make([]*git_model.Branch, 0)
	require.NoError(t, db.GetEngine(db.DefaultContext).Where("repo_id=?", repo.ID).Find(&dbBranches))
	assert.Equalf(t, len(pushedBranches), len(dbBranches), "mismatched number of branches in db")
	dbBranchesMap := make(map[string]*git_model.Branch, len(dbBranches))
	for _, branch := range dbBranches {
		dbBranchesMap[branch.Name] = branch
	}

	deletedBranchesMap := make(map[string]bool, len(deletedBranches))
	for _, branchName := range deletedBranches {
		deletedBranchesMap[branchName] = true
	}

	for _, branchName := range pushedBranches {
		branch, ok := dbBranchesMap[branchName]
		deleted := deletedBranchesMap[branchName]
		assert.True(t, ok, "branch %s not found in database", branchName)
		assert.Equal(t, deleted, branch.IsDeleted, "IsDeleted of %s is %v, but it's expected to be %v", branchName, branch.IsDeleted, deleted)
		commitID, err := gitRepo.GetBranchCommitID(branchName)
		require.NoError(t, err)
		assert.Equal(t, commitID, branch.CommitID)
	}

	require.NoError(t, repo_service.DeleteRepositoryDirectly(db.DefaultContext, user, repo.ID))
}
