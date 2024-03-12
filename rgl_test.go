package rgl_test

import (
	"context"
	"testing"

	"github.com/leighmacdonald/rgl"
	"github.com/leighmacdonald/steamid/v4/steamid"
	"github.com/stretchr/testify/require"
)

func TestAPI(t *testing.T) {
	httpClient := rgl.NewClient()

	t.Run("bans", testBans(httpClient))
	t.Run("profile", testProfile(httpClient))
	t.Run("profiles", testProfiles(httpClient))
	t.Run("profile_teams", testProfileTeams(httpClient))
	t.Run("search_player", testSearchPlayer(httpClient))
	t.Run("match", testMatch(httpClient))
	t.Run("matches", testMatches(httpClient))
	t.Run("team", testTeam(httpClient))
	t.Run("search_team", testSearchTeam(httpClient))
	t.Run("season", testSeason(httpClient))
}

func testBans(client *rgl.LimiterClient) func(t *testing.T) {
	return func(t *testing.T) {
		_, errRange := rgl.Bans(context.Background(), client.Client, 101, 0)
		require.ErrorIs(t, errRange, rgl.ErrOutOfRange)

		bans, errRequest := rgl.Bans(context.Background(), client.Client, 10, 0)
		require.NoError(t, errRequest)
		require.Equal(t, 10, len(bans))
	}
}

func testProfile(client *rgl.LimiterClient) func(t *testing.T) {
	return func(t *testing.T) {
		profile, errRequest := rgl.Profile(context.Background(), client.Client, steamid.New(76561197970669109))
		require.NoError(t, errRequest)
		require.Equal(t, steamid.New(76561197970669109), profile.SteamID)
	}
}

func testProfiles(client *rgl.LimiterClient) func(t *testing.T) {
	return func(t *testing.T) {
		profiles, errRequest := rgl.Profiles(context.Background(), client.Client,
			steamid.Collection{steamid.New(76561197970669109), steamid.New(76561198084134025)})
		require.NoError(t, errRequest)
		require.Equal(t, 2, len(profiles))
	}
}

func testProfileTeams(client *rgl.LimiterClient) func(t *testing.T) {
	return func(t *testing.T) {
		teams, errRequest := rgl.ProfileTeams(context.Background(), client.Client, steamid.New(76561197970669109))
		require.NoError(t, errRequest)
		require.True(t, len(teams) > 30)
	}
}

func testSearchPlayer(client *rgl.LimiterClient) func(t *testing.T) {
	return func(t *testing.T) {
		results, errRequest := rgl.SearchPlayer(context.Background(), client.Client, "camp3r", 100, 0)
		require.NoError(t, errRequest)
		require.True(t, len(results.Results) > 0)
	}
}

func testMatch(client *rgl.LimiterClient) func(t *testing.T) {
	return func(t *testing.T) {
		results, errResults := rgl.Match(context.Background(), client.Client, 1100)
		require.NoError(t, errResults)
		require.Equal(t, "Week 6 - cp_steel", results.MatchName)
	}
}

func testMatches(client *rgl.LimiterClient) func(t *testing.T) {
	return func(t *testing.T) {
		results, errRequest := rgl.Matches(context.Background(), client.Client, 15, 0)
		require.NoError(t, errRequest)
		require.Equal(t, 15, len(results))
	}
}

func testTeam(client *rgl.LimiterClient) func(t *testing.T) {
	return func(t *testing.T) {
		results, errRequest := rgl.Team(context.Background(), client.Client, 7835)
		require.NoError(t, errRequest)
		require.Equal(t, 7835, results.TeamID)
	}
}

func testSearchTeam(client *rgl.LimiterClient) func(t *testing.T) {
	return func(t *testing.T) {
		results, errRequest := rgl.SearchTeam(context.Background(), client.Client, "froyo", 10, 0)
		require.NoError(t, errRequest)
		require.Equal(t, 10, len(results.Results))
	}
}

func testSeason(client *rgl.LimiterClient) func(t *testing.T) {
	return func(t *testing.T) {
		results, errRequest := rgl.Season(context.Background(), client.Client, 50)
		require.NoError(t, errRequest)
		require.Equal(t, results.RegionName, "NA Modern Maps Popup League")
	}
}
