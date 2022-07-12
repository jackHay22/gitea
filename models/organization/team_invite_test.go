// Copyright 2022 The Gitea Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package organization_test

import (
	"testing"

	"code.gitea.io/gitea/models/db"
	"code.gitea.io/gitea/models/organization"
	"code.gitea.io/gitea/models/unittest"
	user_model "code.gitea.io/gitea/models/user"

	"github.com/stretchr/testify/assert"
)

func TestTeam_EmailExists(t *testing.T) {
	assert.NoError(t, unittest.PrepareTestDatabase())

	team := unittest.AssertExistsAndLoadBean(t, &organization.Team{ID: 2}).(*organization.Team)
	user2 := unittest.AssertExistsAndLoadBean(t, &user_model.User{ID: 2}).(*user_model.User)

	// user 2 already added to team 2, should result in error
	_, err := organization.CreateTeamInvite(db.DefaultContext, user2, team, "user2@example.com")
	assert.Error(t, err)
}

func TestTeam_Invite(t *testing.T) {
	assert.NoError(t, unittest.PrepareTestDatabase())

	team := unittest.AssertExistsAndLoadBean(t, &organization.Team{ID: 2}).(*organization.Team)
	user1 := unittest.AssertExistsAndLoadBean(t, &user_model.User{ID: 1}).(*user_model.User)

	_, err := organization.CreateTeamInvite(db.DefaultContext, user1, team, "user3@example.com")
	assert.NoError(t, err)

	// Shouldn't allow duplicate invite
	_, err = organization.CreateTeamInvite(db.DefaultContext, user1, team, "user3@example.com")
	assert.Error(t, err)
}

func TestTeam_RemoveInvite(t *testing.T) {
	assert.NoError(t, unittest.PrepareTestDatabase())

	team := unittest.AssertExistsAndLoadBean(t, &organization.Team{ID: 2}).(*organization.Team)
	user1 := unittest.AssertExistsAndLoadBean(t, &user_model.User{ID: 1}).(*user_model.User)

	invite, err := organization.CreateTeamInvite(db.DefaultContext, user1, team, "user5@example.com")
	assert.NoError(t, err)

	// should remove invite
	assert.NoError(t, organization.RemoveInviteByID(db.DefaultContext, invite.ID, invite.TeamID))

	// invite should not exist
	_, err = organization.GetInviteByToken(db.DefaultContext, invite.Token)
	assert.Error(t, err)
}
