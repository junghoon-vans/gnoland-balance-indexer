package dto

type GraphQLEvent struct {
	Type    string             `json:"type"`
	Func    string             `json:"func"`
	PkgPath string             `json:"pkg_path"`
	Attrs   []GraphQLEventAttr `json:"attrs"`
}

type GraphQLEventAttr struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}
