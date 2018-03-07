package models

type Pipeline struct {
	Name     string `json:"name"`
	TeamName string `json:"team_name"`
}

type Job struct {
	Name          string `json:"name"`
	NextBuild     *Build `json:"next_build"`
	FinishedBuild *Build `json:"finished_build"`
}

type Build struct {
	Status string `json:"status"`
}
