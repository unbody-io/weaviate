package generateflex

type Params struct {
	Prompt        *string
	Task          *string
	Properties    []string
	Messages      *[]interface{}
	SingleResult  bool
	GroupedResult bool
	Debug         bool
	Options       map[string]interface{}
}
