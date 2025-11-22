package srv

type UserProfile struct {
	Uid        string `json:"uid"`
	Location   string `json:"location,omitempty"`
	StudyId    string `json:"study_id,omitempty"`
	SchoolName string `json:"school_name,omitempty"`
	NickName   string `json:"nick_name,omitempty"`
	AvatarUrl  string `json:"avatar_url,omitempty"`
}
