package systemaccesspermission

type SystemAccess struct {
	Report Report `json:"report"`
}
type Title struct {
	Th string `json:"th"`
	En string `json:"en"`
	My string `json:"my"`
}
type Name struct {
	Th string `json:"th"`
	En string `json:"en"`
	My string `json:"my"`
}
type Company struct {
	Code string `json:"code"`
	Name Name   `json:"name"`
}
type Group struct {
	Code string `json:"code"`
	Name Name   `json:"name"`
}
type Menus struct {
	Code string `json:"code"`
	Name Name   `json:"name"`
	View bool   `json:"view"`
}
type Report struct {
	Title   Title   `json:"title"`
	Company Company `json:"company"`
	Group   Group   `json:"group"`
	Menus   []Menus `json:"menus"`
}
