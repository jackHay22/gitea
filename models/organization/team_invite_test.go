// Copyright 2022 The Gitea Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package organization_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"code.gitea.io/gitea/models/db"
	"code.gitea.io/gitea/models/organization"
	"code.gitea.io/gitea/models/unittest"
	user_model "code.gitea.io/gitea/models/user"
)

func TestTeam_EmailExists(t *testing.T) {
	assert.NoError(t, unittest.PrepareTestDatabase())

	team := unittest.AssertExistsAndLoadBean(t, &organization.Team{ID: 2}).(*organization.Team)
	user2 := unittest.AssertExistsAndLoadBean(t, &user_model.User{ID: 2}).(*user_model.User)

	// user 2 already added to team 2, should result in error
	_, err := organization.CreateTeamInvite(db.DefaultContext, user2, team, "user2@example.com")
	assert.Error(t, err)
}
