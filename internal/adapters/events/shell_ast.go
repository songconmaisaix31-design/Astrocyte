package events

type shellASTCommand struct {
	Head    string `json:"head"`
	Dynamic bool   `json:"dynamic"`
}
type shellASTResult struct {
	Commands    []shellASTCommand `json:"commands"`
	ParseErrors int               `json:"parse_errors"`
	ErrorIDs    []string          `json:"error_ids"`
}
