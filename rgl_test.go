package rgl_test

import (
	"context"
	"testing"

	"github.com/leighmacdonald/rgl"
	"github.com/leighmacdonald/steamid/v3/steamid"
	"github.com/stretchr/testify/require"
)

func TestBans(t *testing.T) {
	_, errRange := rgl.Bans(context.Background(), 101, 0)
	require.ErrorIs(t, errRange, rgl.ErrOutOfRange)

	bans, errRequest := rgl.Bans(context.Background(), 10, 0)
	require.NoError(t, errRequest)
	require.Equal(t, 10, len(bans))
}

func TestProfile(t *testing.T) {
	profile, errRequest := rgl.Profile(context.Background(), steamid.New(76561197970669109))
	require.NoError(t, errRequest)
	require.Equal(t, steamid.New(76561197970669109), profile.SteamID)
}

func TestProfiles(t *testing.T) {
	profiles, errRequest := rgl.Profiles(context.Background(),
		steamid.Collection{steamid.New(76561197970669109), steamid.New(76561198084134025)})
	require.NoError(t, errRequest)
	require.Equal(t, 2, len(profiles))
}

func TestProfileTeams(t *testing.T) {
	teams, errRequest := rgl.ProfileTeams(context.Background(), steamid.New(76561197970669109))
	require.NoError(t, errRequest)
	require.True(t, len(teams) > 30)
}

func TestSearchPlayer(t *testing.T) {
	results, errRequest := rgl.SearchPlayer(context.Background(), "camp3r", 100, 0)
	require.NoError(t, errRequest)
	require.True(t, len(results.Results) > 0)
}

func TestMatch(t *testing.T) {
	results, errResults := rgl.Match(context.Background(), 1100)
	require.NoError(t, errResults)
	require.Equal(t, "Week 6 - cp_steel", results.MatchName)
}

func TestMatches(t *testing.T) {
	results, errRequest := rgl.Matches(context.Background(), 15, 0)
	require.NoError(t, errRequest)
	require.Equal(t, 15, len(results))
}

func TestTeam(t *testing.T) {
	results, errRequest := rgl.Team(context.Background(), 7835)
	require.NoError(t, errRequest)
	require.Equal(t, 7835, results.TeamID)
}

func TestSearchTeam(t *testing.T) {
	results, errRequest := rgl.SearchTeam(context.Background(), "froyo", 10, 0)
	require.NoError(t, errRequest)
	require.Equal(t, 10, len(results.Results))
}

func TestSeason(t *testing.T) {
	results, errRequest := rgl.Season(context.Background(), 50)
	require.NoError(t, errRequest)
	require.Equal(t, results.RegionName, "NA Modern Maps Popup League")
}
