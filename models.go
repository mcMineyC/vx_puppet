package vx_puppet

type GradeInfo struct {
	Class       string `json:"class"`
	GradeLetter string `json:"gradeLetter"`
	Grade       string `json:"grade"`
	NewUpdates  int    `json:"new_updates"`
}
