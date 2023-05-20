package rgl

import (
	"context"
	"github.com/leighmacdonald/steamid/v2/steamid"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestBans(t *testing.T) {
	_, errRange := Bans(context.Background(), 101, 0)
	require.ErrorIs(t, errRange, ErrOutOfRange)
	bans, errRequest := Bans(context.Background(), 10, 0)
	require.NoError(t, errRequest)
	require.Equal(t, 10, len(bans))
}

func TestProfile(t *testing.T) {
	profile, errRequest := Profile(context.Background(), 76561197970669109)
	require.NoError(t, errRequest)
	require.Equal(t, steamid.SID64(76561197970669109), profile.SteamId)
}

func TestProfiles(t *testing.T) {
	profiles, errRequest := Profiles(context.Background(), steamid.Collection{76561197970669109, 76561198084134025})
	require.NoError(t, errRequest)
	require.Equal(t, 2, len(profiles))
}

func TestProfileTeams(t *testing.T) {
	teams, errRequest := ProfileTeams(context.Background(), 76561197970669109)
	require.NoError(t, errRequest)
	require.True(t, len(teams) > 30)
}

func TestSearchPlayer(t *testing.T) {
	results, errRequest := SearchPlayer(context.Background(), "camp3r", 100, 0)
	require.NoError(t, errRequest)
	require.True(t, len(results.Results) > 0)
}

func TestMatch(t *testing.T) {
	results, errResults := Match(context.Background(), 1100)
	require.NoError(t, errResults)
	require.Equal(t, "Week 6 - cp_steel", results.MatchName)
}

func TestMatches(t *testing.T) {
	results, errRequest := Matches(context.Background(), 15, 0)
	require.NoError(t, errRequest)
	require.Equal(t, 15, len(results))
}

func TestTeam(t *testing.T) {
	results, errRequest := Team(context.Background(), 7835)
	require.NoError(t, errRequest)
	require.Equal(t, 7835, results.TeamId)
}

func TestSearchTeam(t *testing.T) {
	results, errRequest := SearchTeam(context.Background(), "froyo", 10, 0)
	require.NoError(t, errRequest)
	require.Equal(t, 10, len(results.Results))
}

func TestSeason(t *testing.T) {
	results, errRequest := Season(context.Background(), 50)
	require.NoError(t, errRequest)
	require.Equal(t, results.RegionName, "NA Modern Maps Popup League")
}
