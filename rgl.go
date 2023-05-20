package rgl

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/leighmacdonald/steamid/v2/steamid"
	"github.com/pkg/errors"
	"io"
	"net/http"
	"net/url"
	"time"
)

const (
	baseUrl = "https://api.rgl.gg/v0"
)

func call(ctx context.Context, method string, fullUrl string, body any, receiver any) error {
	var reqBody io.Reader
	if body != nil {
		rb, errMarshal := json.Marshal(body)
		if errMarshal != nil {
			return errMarshal
		}
		reqBody = bytes.NewReader(rb)
	}

	req, errReq := http.NewRequestWithContext(ctx, method, fullUrl, reqBody)
	if errReq != nil {
		return errors.Wrap(errReq, "Failed to create request")
	}
	req.Header.Add("Content-Type", `application/json`)
	resp, errResp := client.Do(ctx, req)
	if errResp != nil {
		return errors.Wrap(errResp, "Failed to call endpoint")
	}
	// TODO influence rate limit bucket?
	//rateLimit = resp.Header.Get("X-Ratelimit-Limit")
	//rateRemaining = resp.Header.Get("X-Ratelimit-Remaining")
	//rateReset = resp.Header.Get("X-Ratelimit-Reset")
	respBody, errRead := io.ReadAll(resp.Body)
	if errRead != nil {
		return errors.Wrap(errRead, "Failed to read response body")
	}
	defer func() {
		_ = resp.Body.Close()
	}()
	if errJson := json.Unmarshal(respBody, &receiver); errJson != nil {
		return errors.Wrap(errJson, "Failed to unmarshal json payload")
	}
	return nil
}

type Ban struct {
	SteamId   steamid.SID64 `json:"steamId,float64"`
	Alias     string        `json:"alias"`
	ExpiresAt time.Time     `json:"expiresAt"`
	CreatedAt time.Time     `json:"createdAt"`
	Reason    string        `json:"reason"`
}

var ErrOutOfRange = errors.New("Value out of range")

func validateQuery(take int, skip int) error {
	if take > 100 || take < 0 || skip < 0 {
		return ErrOutOfRange
	}
	return nil
}

func Bans(ctx context.Context, take int, skip int) ([]Ban, error) {
	if errValidate := validateQuery(take, skip); errValidate != nil {
		return nil, errValidate
	}
	var bans []Ban
	errBans := call(ctx, http.MethodGet, mkPagedPath("/bans/paged", take, skip), nil, &bans)
	if errBans != nil {
		return nil, errBans
	}
	return bans, nil
}

type PlayerStatus struct {
	IsVerified    bool `json:"isVerified"`
	IsBanned      bool `json:"isBanned"`
	IsOnProbation bool `json:"isOnProbation"`
}

type PlayerTeams struct {
	Sixes      *PlayerTeam `json:"sixes"`
	Highlander *PlayerTeam `json:"highlander"`
	Prolander  *PlayerTeam `json:"prolander"`
}

type PlayerTeam struct {
	Id           int    `json:"id"`
	Tag          string `json:"tag"`
	Name         string `json:"name"`
	Status       string `json:"status"`
	SeasonId     int    `json:"seasonId"`
	DivisionId   int    `json:"divisionId"`
	DivisionName string `json:"divisionName"`
}

type PlayerBanInformation struct {
	EndsAt time.Time `json:"endsAt"`
	Reason string    `json:"reason"`
}

type Player struct {
	SteamId        steamid.SID64         `json:"steamId,float64"`
	Avatar         string                `json:"avatar"`
	Name           string                `json:"name"`
	UpdatedAt      time.Time             `json:"updatedAt"`
	Status         PlayerStatus          `json:"status"`
	BanInformation *PlayerBanInformation `json:"banInformation"`
	CurrentTeams   PlayerTeams           `json:"currentTeams"`
}

func Profile(ctx context.Context, sid64 steamid.SID64) (*Player, error) {
	var player Player
	errProfile := call(ctx, http.MethodGet, mkPath(fmt.Sprintf("/profile/%d", sid64)), nil, &player)
	if errProfile != nil {
		return nil, errProfile
	}
	return &player, nil
}

type TeamStats struct {
	Wins         int `json:"wins"`
	WinsWithout  int `json:"winsWithout"`
	Loses        int `json:"loses"`
	LosesWithout int `json:"losesWithout"`
	GamesPlayed  int `json:"gamesPlayed"`
	GamesWithout int `json:"gamesWithout"`
}

type ProfileTeam struct {
	FormatId     int       `json:"formatId"`
	FormatName   string    `json:"formatName"`
	RegionId     int       `json:"regionId"`
	RegionName   string    `json:"regionName"`
	SeasonId     int       `json:"seasonId"`
	SeasonName   string    `json:"seasonName"`
	StartedAt    time.Time `json:"startedAt"`
	DivisionId   int       `json:"divisionId"`
	DivisionName string    `json:"divisionName"`
	LeftAt       time.Time `json:"leftAt"`
	TeamName     string    `json:"teamName"`
	TeamTag      string    `json:"teamTag"`
	TeamId       int       `json:"teamId"`
	Stats        TeamStats `json:"stats"`
}

func ProfileTeams(ctx context.Context, sid64 steamid.SID64) ([]ProfileTeam, error) {
	var teams []ProfileTeam
	errProfile := call(ctx, http.MethodGet, mkPath(fmt.Sprintf("/profile/%d/teams", sid64)), nil, &teams)
	if errProfile != nil {
		return nil, errProfile
	}
	return teams, nil
}

func Profiles(ctx context.Context, steamIIds steamid.Collection) ([]*Player, error) {
	if len(steamIIds) > 100 || len(steamIIds) == 0 {
		return nil, ErrOutOfRange
	}
	var players []*Player
	if errPlayers := call(ctx, http.MethodPost, mkPath("/profile/getmany"), steamIIds.ToStringSlice(), &players); errPlayers != nil {
		return nil, errPlayers
	}
	return players, nil
}

type SearchPlayerResults struct {
	Results       steamid.Collection `json:"results,float64"`
	Count         int                `json:"count"`
	TotalHitCount int                `json:"totalHitCount"`
}

type searchNameRequest struct {
	NameContains string `json:"nameContains"`
}

func mkPagedPath(path string, take int, skip int) string {
	u, errUrl := url.Parse(baseUrl + path)
	if errUrl != nil {
		panic(errUrl)
	}
	values := u.Query()
	values.Set("take", fmt.Sprintf("%d", take))
	values.Set("skip", fmt.Sprintf("%d", skip))
	u.RawQuery = values.Encode()
	return u.String()
}

func mkPath(path string) string {
	u, errUrl := url.Parse(baseUrl + path)
	if errUrl != nil {
		panic(errUrl)
	}
	return u.String()
}

func SearchPlayer(ctx context.Context, name string, take int, skip int) (*SearchPlayerResults, error) {
	if name == "" {
		return nil, ErrOutOfRange
	}
	if errValidate := validateQuery(take, skip); errValidate != nil {
		return nil, errValidate
	}
	var results SearchPlayerResults
	if errSearch := call(ctx, http.MethodPost, mkPagedPath("/search/players", take, skip),
		searchNameRequest{NameContains: name}, &results); errSearch != nil {
		return nil, errSearch
	}
	return &results, nil
}

type MatchTeam struct {
	TeamName string `json:"teamName"`
	TeamTag  string `json:"teamTag"`
	TeamId   int    `json:"teamId"`
	IsHome   bool   `json:"isHome"`
	Points   string `json:"points"`
}

type MatchMaps struct {
	MapName   string `json:"mapName"`
	HomeScore int    `json:"homeScore"`
	AwayScore int    `json:"awayScore"`
}

type MatchOverview struct {
	MatchId      int         `json:"matchId"`
	SeasonName   string      `json:"seasonName"`
	DivisionName string      `json:"divisionName"`
	DivisionId   int         `json:"divisionId"`
	SeasonId     int         `json:"seasonId"`
	MatchDate    time.Time   `json:"matchDate"`
	MatchName    string      `json:"matchName"`
	IsForfeit    bool        `json:"isForfeit"`
	Winner       int         `json:"winner"`
	Teams        []MatchTeam `json:"teams"`
	Maps         []MatchMaps `json:"maps"`
}

func Match(ctx context.Context, matchId int64) (*MatchOverview, error) {
	if matchId <= 0 {
		return nil, ErrOutOfRange
	}
	var match MatchOverview
	if errSearch := call(ctx, http.MethodGet, mkPath(fmt.Sprintf("/matches/%d", matchId)), nil, &match); errSearch != nil {
		return nil, errSearch
	}
	return &match, nil
}

type emptyReq struct{}

func Matches(ctx context.Context, take int, skip int) ([]*MatchOverview, error) {
	if errValidate := validateQuery(take, skip); errValidate != nil {
		return nil, errValidate
	}
	var matches []*MatchOverview
	if errSearch := call(ctx, http.MethodPost, mkPagedPath("/matches/paged", take, skip), emptyReq{}, &matches); errSearch != nil {
		return nil, errSearch
	}
	return matches, nil
}

type TeamPlayer struct {
	Name     string    `json:"name"`
	SteamId  string    `json:"steamId"`
	IsLeader bool      `json:"isLeader"`
	JoinedAt time.Time `json:"joinedAt"`
	LeftAt   time.Time `json:"leftAt"`
}

type TeamOverview struct {
	TeamId       int          `json:"teamId"`
	LinkedTeams  []int        `json:"linkedTeams"`
	SeasonId     int          `json:"seasonId"`
	DivisionId   int          `json:"divisionId"`
	DivisionName string       `json:"divisionName"`
	TeamLeader   string       `json:"teamLeader"`
	CreatedAt    time.Time    `json:"createdAt"`
	UpdatedAt    time.Time    `json:"updatedAt"`
	Tag          string       `json:"tag"`
	Name         string       `json:"name"`
	FinalRank    int          `json:"finalRank"`
	Players      []TeamPlayer `json:"players"`
}

func Team(ctx context.Context, teamId int64) (*TeamOverview, error) {
	if teamId <= 0 {
		return nil, ErrOutOfRange
	}
	var team TeamOverview
	if errCall := call(ctx, http.MethodGet, mkPath(fmt.Sprintf("/teams/%d", teamId)), nil, &team); errCall != nil {
		return nil, errCall
	}
	return &team, nil
}

type SearchTeamResults struct {
	Results       []string `json:"results"`
	Count         int      `json:"count"`
	TotalHitCount int      `json:"totalHitCount"`
}

func SearchTeam(ctx context.Context, name string, take int, skip int) (*SearchTeamResults, error) {
	if name == "" {
		return nil, ErrOutOfRange
	}
	if errValidate := validateQuery(take, skip); errValidate != nil {
		return nil, errValidate
	}
	var results SearchTeamResults
	if errSearch := call(ctx, http.MethodPost, mkPagedPath("/search/teams", take, skip),
		searchNameRequest{NameContains: name}, &results); errSearch != nil {
		return nil, errSearch
	}
	return &results, nil
}

type SeasonOverview struct {
	Name                      string         `json:"name"`
	DivisionSorting           map[string]int `json:"divisionSorting"`
	FormatName                string         `json:"formatName"`
	RegionName                string         `json:"regionName"`
	Maps                      []string       `json:"maps"`
	ParticipatingTeams        []int          `json:"participatingTeams"`
	MatchesPlayedDuringSeason []int          `json:"matchesPlayedDuringSeason"`
}

func Season(ctx context.Context, seasonId int64) (*SeasonOverview, error) {
	if seasonId <= 0 {
		return nil, ErrOutOfRange
	}
	var season SeasonOverview
	if errCall := call(ctx, http.MethodGet, mkPath(fmt.Sprintf("/seasons/%d", seasonId)), nil, &season); errCall != nil {
		return nil, errCall
	}
	return &season, nil
}
