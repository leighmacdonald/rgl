package rgl

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/leighmacdonald/steamid/v4/steamid"
)

const (
	baseURL       = "https://api.rgl.gg/v0"
	maxQueryCount = 100
)

var (
	ErrRateLimit      = errors.New("rate limited (429)")
	ErrRequestClose   = errors.New("failed to close request body")
	ErrRequestEncode  = errors.New("failed to marshal request payload")
	ErrRequestDecode  = errors.New("failed to unmarshal json response payload")
	ErrRequestCreate  = errors.New("failed to make request")
	ErrRequestPerform = errors.New("failed to call endpoint")
	ErrRequestStatus  = errors.New("invalid status code")
)

type HTTPExecutor interface {
	Do(req *http.Request) (*http.Response, error)
}

func call(ctx context.Context, httpClient HTTPExecutor, method string, fullURL string, body any, receiver any) error {
	var reqBody io.Reader

	if body != nil {
		rb, errMarshal := json.Marshal(body)
		if errMarshal != nil {
			return errors.Join(errMarshal, ErrRequestEncode)
		}

		reqBody = bytes.NewReader(rb)
	}

	req, errReq := http.NewRequestWithContext(ctx, method, fullURL, reqBody)
	if errReq != nil {
		return errors.Join(errReq, ErrRequestCreate)
	}

	req.Header.Set("User-Agent", "bd-api/1.0")
	req.Header.Add("Content-Type", `application/json`)

	resp, errResp := httpClient.Do(req)
	if errResp != nil {
		return errors.Join(errResp, ErrRequestPerform)
	}

	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode == http.StatusTooManyRequests {
		return ErrRateLimit
	}

	if !(resp.StatusCode >= http.StatusOK && resp.StatusCode <= http.StatusIMUsed) {
		return fmt.Errorf("%w: %s", ErrRequestStatus, resp.Status)
	}

	if errJSON := json.NewDecoder(resp.Body).Decode(&receiver); errJSON != nil {
		return errors.Join(errJSON, ErrRequestDecode)
	}

	if errClose := resp.Body.Close(); errClose != nil {
		return errors.Join(errClose, ErrRequestClose)
	}

	return nil
}

type Ban struct {
	SteamID   string    `json:"steamId"`
	Alias     string    `json:"alias"`
	ExpiresAt time.Time `json:"expiresAt"`
	CreatedAt time.Time `json:"createdAt"`
	Reason    string    `json:"reason"`
}

var ErrOutOfRange = errors.New("value out of range")

func validateQuery(take int, skip int) error {
	if take > maxQueryCount || take < 0 || skip < 0 {
		return ErrOutOfRange
	}

	return nil
}

func Bans(ctx context.Context, httpClient HTTPExecutor, take int, skip int) ([]Ban, error) {
	if errValidate := validateQuery(take, skip); errValidate != nil {
		return nil, errValidate
	}

	var bans []Ban

	errBans := call(ctx, httpClient, http.MethodGet, mkPagedPath("/bans/paged", take, skip), nil, &bans)
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
	ID           int    `json:"id"`
	Tag          string `json:"tag"`
	Name         string `json:"name"`
	Status       string `json:"status"`
	SeasonID     int    `json:"seasonId"`
	DivisionID   int    `json:"divisionId"`
	DivisionName string `json:"divisionName"`
}

type PlayerBanInformation struct {
	EndsAt time.Time `json:"endsAt"`
	Reason string    `json:"reason"`
}

type Player struct {
	SteamID        steamid.SteamID       `json:"steamId"`
	Avatar         string                `json:"avatar"`
	Name           string                `json:"name"`
	UpdatedAt      time.Time             `json:"updatedAt"`
	Status         PlayerStatus          `json:"status"`
	BanInformation *PlayerBanInformation `json:"banInformation"`
	CurrentTeams   PlayerTeams           `json:"currentTeams"`
}

func Profile(ctx context.Context, httpClient HTTPExecutor, steamID steamid.SteamID) (*Player, error) {
	var player Player
	if errProfile := call(ctx, httpClient, http.MethodGet,
		mkPath(fmt.Sprintf("/profile/%d", steamID.Int64())), nil, &player); errProfile != nil {
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
	FormatID     int       `json:"formatId"`
	FormatName   string    `json:"formatName"`
	RegionID     int       `json:"regionId"`
	RegionName   string    `json:"regionName"`
	SeasonID     int64     `json:"seasonId"`
	SeasonName   string    `json:"seasonName"`
	StartedAt    time.Time `json:"startedAt"`
	DivisionID   int       `json:"divisionId"`
	DivisionName string    `json:"divisionName"`
	LeftAt       time.Time `json:"leftAt"`
	TeamName     string    `json:"teamName"`
	TeamTag      string    `json:"teamTag"`
	TeamID       int       `json:"teamId"`
	Stats        TeamStats `json:"stats"`
}

func ProfileTeams(ctx context.Context, httpClient HTTPExecutor, sid64 steamid.SteamID) ([]ProfileTeam, error) {
	var teams []ProfileTeam

	path := mkPath(fmt.Sprintf("/profile/%d/teams", sid64.Int64()))
	errProfile := call(ctx, httpClient, http.MethodGet, path, nil, &teams)
	if errProfile != nil {
		return nil, errProfile
	}

	return teams, nil
}

func Profiles(ctx context.Context, httpClient HTTPExecutor, steamIIds steamid.Collection) ([]*Player, error) {
	if len(steamIIds) > 100 || len(steamIIds) == 0 {
		return nil, ErrOutOfRange
	}

	var players []*Player
	if errPlayers := call(ctx, httpClient, http.MethodPost,
		mkPath("/profile/getmany"), steamIIds.ToStringSlice(), &players); errPlayers != nil {
		return nil, errPlayers
	}

	return players, nil
}

type SearchPlayerResults struct {
	Results       steamid.Collection `json:"results"`
	Count         int                `json:"count"`
	TotalHitCount int                `json:"totalHitCount"`
}

type searchNameRequest struct {
	NameContains string `json:"nameContains"`
}

func mkPagedPath(path string, take int, skip int) string {
	parsedURL, errURL := url.Parse(baseURL + path)
	if errURL != nil {
		panic(errURL)
	}

	values := parsedURL.Query()
	values.Set("take", fmt.Sprintf("%d", take))
	values.Set("skip", fmt.Sprintf("%d", skip))
	parsedURL.RawQuery = values.Encode()

	return parsedURL.String()
}

func mkPath(path string) string {
	u, errURL := url.Parse(baseURL + path)
	if errURL != nil {
		panic(errURL)
	}

	return u.String()
}

func SearchPlayer(ctx context.Context, httpClient HTTPExecutor, name string, take int, skip int) (*SearchPlayerResults, error) {
	if name == "" {
		return nil, ErrOutOfRange
	}

	if errValidate := validateQuery(take, skip); errValidate != nil {
		return nil, errValidate
	}

	var results SearchPlayerResults
	if errSearch := call(ctx, httpClient, http.MethodPost, mkPagedPath("/search/players", take, skip),
		searchNameRequest{NameContains: name}, &results); errSearch != nil {
		return nil, errSearch
	}

	return &results, nil
}

type MatchTeam struct {
	TeamName string  `json:"teamName"`
	TeamTag  string  `json:"teamTag"`
	TeamID   int     `json:"teamId"`
	IsHome   bool    `json:"isHome"`
	Points   float32 `json:"points,string"`
}

type MatchMaps struct {
	MapName   string `json:"mapName"`
	HomeScore int    `json:"homeScore"`
	AwayScore int    `json:"awayScore"`
}

type MatchOverview struct {
	MatchID      int         `json:"matchId"`
	SeasonName   string      `json:"seasonName"`
	DivisionName string      `json:"divisionName"`
	DivisionID   int         `json:"divisionId"`
	SeasonID     int         `json:"seasonId"`
	MatchDate    time.Time   `json:"matchDate"`
	MatchName    string      `json:"matchName"`
	IsForfeit    bool        `json:"isForfeit"`
	RegionID     int         `json:"regionId"`
	Winner       int         `json:"winner"`
	Teams        []MatchTeam `json:"teams"`
	Maps         []MatchMaps `json:"maps"`
}

func Match(ctx context.Context, httpClient HTTPExecutor, matchID int64) (*MatchOverview, error) {
	if matchID <= 0 {
		return nil, ErrOutOfRange
	}

	var match MatchOverview
	if errSearch := call(ctx, httpClient, http.MethodGet, mkPath(fmt.Sprintf("/matches/%d", matchID)), nil, &match); errSearch != nil {
		return nil, errSearch
	}

	return &match, nil
}

type emptyReq struct{}

func Matches(ctx context.Context, httpClient HTTPExecutor, take int, skip int) ([]*MatchOverview, error) {
	if errValidate := validateQuery(take, skip); errValidate != nil {
		return nil, errValidate
	}

	var matches []*MatchOverview
	if errSearch := call(ctx, httpClient, http.MethodPost,
		mkPagedPath("/matches/paged", take, skip), emptyReq{}, &matches); errSearch != nil {
		return nil, errSearch
	}

	return matches, nil
}

type TeamPlayer struct {
	Name      string          `json:"name"`
	SteamID   steamid.SteamID `json:"steamId"`
	IsLeader  bool            `json:"isLeader"`
	JoinedAt  time.Time       `json:"joinedAt"`
	LeftAt    *time.Time      `json:"leftAt"`
	CreatedOn time.Time       `json:"created_on"`
	UpdatedOn time.Time       `json:"updatedOn"`
}

type TeamOverview struct {
	TeamID       int          `json:"teamId"`
	LinkedTeams  []int        `json:"linkedTeams"`
	SeasonID     int          `json:"seasonId"`
	DivisionID   int          `json:"divisionId"`
	DivisionName string       `json:"divisionName"`
	TeamLeader   string       `json:"teamLeader"`
	CreatedAt    time.Time    `json:"createdAt"`
	UpdatedAt    time.Time    `json:"updatedAt"`
	Tag          string       `json:"tag"`
	Name         string       `json:"name"`
	FinalRank    int          `json:"finalRank"`
	Players      []TeamPlayer `json:"players"`
}

func Team(ctx context.Context, httpClient HTTPExecutor, teamID int64) (*TeamOverview, error) {
	if teamID <= 0 {
		return nil, ErrOutOfRange
	}

	var team TeamOverview
	if errCall := call(ctx, httpClient, http.MethodGet, mkPath(fmt.Sprintf("/teams/%d", teamID)), nil, &team); errCall != nil {
		return nil, errCall
	}

	return &team, nil
}

type SearchTeamResults struct {
	Results       []string `json:"results"`
	Count         int      `json:"count"`
	TotalHitCount int      `json:"totalHitCount"`
}

func SearchTeam(ctx context.Context, httpClient HTTPExecutor, name string, take int, skip int) (*SearchTeamResults, error) {
	if name == "" {
		return nil, ErrOutOfRange
	}

	if errValidate := validateQuery(take, skip); errValidate != nil {
		return nil, errValidate
	}

	var results SearchTeamResults
	if errSearch := call(ctx, httpClient, http.MethodPost, mkPagedPath("/search/teams", take, skip),
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

// Season fetched and returns a SeasonOverview containing high level info about the season as a whole.
func Season(ctx context.Context, httpClient HTTPExecutor, seasonID int64) (*SeasonOverview, error) {
	if seasonID <= 0 {
		return nil, ErrOutOfRange
	}

	var season SeasonOverview
	if errCall := call(ctx, httpClient, http.MethodGet, mkPath(fmt.Sprintf("/seasons/%d", seasonID)), nil, &season); errCall != nil {
		return nil, errCall
	}

	return &season, nil
}
