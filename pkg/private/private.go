package private

//nolint:gochecknoglobals
var (
	kind = "custom"
	from = "internal"
)

type Output struct {
	Kind string `json:"kind"`
	From string `json:"from"`
}

func Info() any {
	return &Output{Kind: kind, From: from}
}

func MD5() string {
	return kind
}
