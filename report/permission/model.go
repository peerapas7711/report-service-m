package permission

type Report struct {
	Title string
	Group string
	User  User
}

type User struct {
	Username  string
	Programs  []CheckItem
	Companies []RowCheck
	EmpTypes  []RowCheck
}

type CheckItem struct {
	Label   string
	Checked bool
}

type RowCheck struct {
	Code    string
	Name    string
	Checked bool
}
