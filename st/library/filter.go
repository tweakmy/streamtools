package library

import (
	"github.com/nytlabs/gojee"
	"github.com/nytlabs/streamtools/st/blocks" // blocks
)

type Filter struct {
	blocks.Block
	queryrule chan chan interface{}
	inrule    chan interface{}
	in        chan interface{}
	out       chan interface{}
	quit      chan interface{}
}

// a bit of boilerplate for streamtools
func NewFilter() blocks.BlockInterface {
	return &Filter{}
}

func (b *Filter) Setup() {
	b.Kind = "Filter"
	b.in = b.InRoute("in")
	b.inrule = b.InRoute("rule")
	b.queryrule = b.QueryRoute("rule")
	b.quit = b.Quit()
	b.out = b.Broadcast()
}

func (b *Filter) Run() {
	var filter string
	var parsed *jee.TokenTree

	for {
		select {
		case msg := <-b.in:
			if parsed == nil {
				b.Error("no filter set")
				break
			}

			e, err := jee.Eval(parsed, msg)
			if err != nil {
				b.Error(err)
				break
			}

			eval, ok := e.(bool)
			if !ok {
				break
			}

			if eval == true {
				b.out <- msg
			}

		case ruleI := <-b.inrule:
			rule := ruleI.(map[string]interface{})
			filterS, ok := rule["Filter"].(string)
			if !ok {
				b.Error("bad filter")
				break
			}

			lexed, err := jee.Lexer(filterS)
			if err != nil {
				b.Error(err)
				break
			}

			tree, err := jee.Parser(lexed)
			if err != nil {
				b.Error(err)
				break
			}

			parsed = tree
			filter = filterS

		case c := <-b.queryrule:
			// deal with a query request
			c <- map[string]string{
				"Filter": filter,
			}
		case <-b.quit:
			// quit the block
			return
		}
	}
}
