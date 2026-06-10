package spider

type ZreadFetchResult struct {
	Code int          `json:"code"`
	Msg  string       `json:"msg"`
	Data []ZreadGroup `json:"data"`
}

type ZreadGroup struct {
	Title    string      `json:"title"`
	TimeSpan ZreadTime   `json:"time_span"`
	Repos    []ZreadRepo `json:"repos"`
}

type ZreadTime struct {
	Start string `json:"start"` // MM/DD
	End   string `json:"end"`   // MM/DD
}

type ZreadRepo struct {
	RepoID        string   `json:"repo_id"`
	Owner         string   `json:"owner"`
	Name          string   `json:"name"`
	URL           string   `json:"url"`
	Description   string   `json:"description"`
	DescriptionZh string   `json:"description_zh"`
	StarCount     int      `json:"star_count"`
	Language      string   `json:"language"`
	Topics        []string `json:"topics"`
	WikiID        string   `json:"wiki_id"`
	CreatedAt     int64    `json:"created_at"`
	UpdatedAt     int64    `json:"updated_at"`
}
