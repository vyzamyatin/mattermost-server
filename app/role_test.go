// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package app

import (
	"fmt"
	"testing"

	"github.com/mattermost/mattermost-server/v5/model"
	"github.com/mattermost/mattermost-server/v5/utils/slices"
	"github.com/stretchr/testify/require"
)

// also tests (*App).ChannelRolesGrantPermission and (*App).GetRoleByName
func TestGetRolesByNames(t *testing.T) {
	th := Setup(t).InitBasic()
	defer th.TearDown()

	th.App.SetLicense(model.NewTestLicense(""))
	th.App.SetPhase2PermissionsMigrationStatus(true)

	permissionsDefault := []string{
		model.PERMISSION_MANAGE_CHANNEL_ROLES.Id,
		model.PERMISSION_MANAGE_PUBLIC_CHANNEL_MEMBERS.Id,
	}

	// Defer resetting the system scheme permissions
	systemSchemeRoles, err := th.App.GetRolesByNames([]string{
		model.CHANNEL_GUEST_ROLE_ID,
		model.CHANNEL_USER_ROLE_ID,
		model.CHANNEL_ADMIN_ROLE_ID,
	})
	require.Nil(t, err)
	require.Len(t, systemSchemeRoles, 3)

	// defer resetting the system role permissions
	for _, systemRole := range systemSchemeRoles {
		defer th.App.PatchRole(systemRole, &model.RolePatch{
			Permissions: &systemRole.Permissions,
		})
	}

	// Make a channel scheme, clear its permissions
	channelScheme, err := th.App.CreateScheme(&model.Scheme{
		Name:        model.NewId(),
		DisplayName: model.NewId(),
		Scope:       model.SCHEME_SCOPE_CHANNEL,
	})
	require.Nil(t, err)
	defer th.App.DeleteScheme(channelScheme.Id)

	team := th.CreateTeam()
	defer th.App.PermanentDeleteTeamId(team.Id)

	// Make a channel, set its scheme
	channel := th.CreateChannel(team)
	channel.SchemeId = &channelScheme.Id
	channel, err = th.App.UpdateChannelScheme(channel)
	require.Nil(t, err)

	// See tests/channel-role-has-permission.md for an explanation of the below matrix.
	truthTable := [][]bool{
		{true, true, true, true},
		{true, true, false, false},
		{true, false, true, true},
		{true, false, false, true},
		{false, true, true, false},
		{false, true, false, false},
		{false, false, true, false},
		{false, false, false, false},
	}

	test := func(higherScopedGuestRoleName, higherScopedUserRoleName, higherScopedAdminRoleName string) {
		for _, roleNameUnderTest := range []string{higherScopedGuestRoleName, higherScopedUserRoleName, higherScopedAdminRoleName} {
			for i, row := range truthTable {
				p, q, r, shouldHavePermission := row[0], row[1], row[2], row[3]

				// select the permission to test (moderated or non-moderated)
				var permission *model.Permission
				if q {
					permission = model.PERMISSION_CREATE_POST // moderated
				} else {
					permission = model.PERMISSION_READ_CHANNEL // non-moderated
				}

				// add or remove the permission from the higher-scoped scheme
				higherScopedRole, testErr := th.App.GetRoleByName(roleNameUnderTest)
				require.Nil(t, testErr)

				var higherScopedPermissions []string
				if p {
					higherScopedPermissions = []string{permission.Id}
				} else {
					higherScopedPermissions = permissionsDefault
				}
				higherScopedRole, testErr = th.App.PatchRole(higherScopedRole, &model.RolePatch{Permissions: &higherScopedPermissions})
				require.Nil(t, testErr)

				// get channel role
				var channelRoleName string
				switch roleNameUnderTest {
				case higherScopedGuestRoleName:
					channelRoleName = channelScheme.DefaultChannelGuestRole
				case higherScopedUserRoleName:
					channelRoleName = channelScheme.DefaultChannelUserRole
				case higherScopedAdminRoleName:
					channelRoleName = channelScheme.DefaultChannelAdminRole
				}
				channelRole, testErr := th.App.GetRoleByName(channelRoleName)
				require.Nil(t, testErr)

				// add or remove the permission from the channel scheme
				var channelSchemePermissions []string
				if r {
					channelSchemePermissions = []string{permission.Id}
				} else {
					channelSchemePermissions = permissionsDefault
				}
				channelRole, testErr = th.App.PatchRole(channelRole, &model.RolePatch{Permissions: &channelSchemePermissions})
				require.Nil(t, testErr)

				// tests
				actualRoles, testErr := th.App.GetRolesByNames([]string{channelRole.Name})
				require.Nil(t, testErr)
				require.Len(t, actualRoles, 1)

				actualRole := actualRoles[0]
				require.NotNil(t, actualRole)
				require.Equal(t, channelRole.Name, actualRole.Name)

				failMsg := fmt.Sprintf("* higher-scoped role name: %s\n* truth table row: %d\n* actual permissions: %v\n* permission under test: %s\n* expecting: %v\n* role display name: %s\n* role name: %s", higherScopedRole.Name, i+1, actualRole.Permissions, permission.Id, shouldHavePermission, actualRole.DisplayName, actualRole.Name)
				require.Equal(t, shouldHavePermission, slices.IncludesString(actualRole.Permissions, permission.Id), failMsg)

				// Test authorization while we're at it
				require.Equal(t, shouldHavePermission, th.App.ChannelRolesGrantPermission([]string{channelRole.Name}, permission.Id, channel.Id))

				// ensure that the higher-scoped roles remain unchanged.
				require.Equal(t, p, th.App.ChannelRolesGrantPermission([]string{higherScopedRole.Name}, permission.Id, channel.Id))

				// also test GetRoleByName
				actualRole2, testErr := th.App.GetRoleByName(channelRole.Name)
				require.Nil(t, testErr)
				require.Equal(t, actualRole.Permissions, actualRole2.Permissions)
			}
		}
	}

	// test all 24 permutations where the higher-scoped scheme is the system scheme
	test(model.CHANNEL_GUEST_ROLE_ID, model.CHANNEL_USER_ROLE_ID, model.CHANNEL_ADMIN_ROLE_ID)

	// create a team scheme, assign it to the team
	teamScheme, err := th.App.CreateScheme(&model.Scheme{
		Name:        model.NewId(),
		DisplayName: model.NewId(),
		Scope:       model.SCHEME_SCOPE_TEAM,
	})
	require.Nil(t, err)
	defer th.App.DeleteScheme(teamScheme.Id)
	team.SchemeId = &teamScheme.Id
	team, err = th.App.UpdateTeamScheme(team)
	require.Nil(t, err)

	// test all 24 permutations where the higher-scoped scheme is a team scheme
	test(teamScheme.DefaultChannelGuestRole, teamScheme.DefaultChannelUserRole, teamScheme.DefaultChannelAdminRole)
}
