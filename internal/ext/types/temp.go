package types

import "time"

// TempData は子プロセス起動時の状態を保持する一時ファイルデータ。
// JSON でシリアライズして tmp ファイルに保存する。
type TempData struct {
	ParentPID         int      `json:"parent_pid"`
	ChildPID          int      `json:"child_pid"`
	ConfigFile        string   `json:"config_file"`
	Program           string   `json:"program"`
	ProgramArgs       []string `json:"program_args"`
	Tag               string   `json:"tag"`
	Imports           []string `json:"imports"`
	WaitTimeout       int64    `json:"wait_timeout"` // seconds
	HideWindow        bool     `json:"hide_window"`
	DeleterPath       string   `json:"deleter_path"`
	DeleterHideWindow bool     `json:"deleter_hide_window"`
}

// WaitDuration は WaitTimeout を time.Duration に変換して返す。
func (t *TempData) WaitDuration() time.Duration {
	return time.Duration(t.WaitTimeout) * time.Second
}
