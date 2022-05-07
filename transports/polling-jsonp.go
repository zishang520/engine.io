package transports

import (
	"encoding/json"
	"github.com/zishang520/engine.io/packet"
	"github.com/zishang520/engine.io/types"
	"github.com/zishang520/engine.io/utils"
	"net/url"
	"regexp"
)

const (
	rDoubleSlashes string = `\\\\n`
	rSlashes       string = `(\\)?\\n`
)

type jsonp struct {
	*polling

	head string
	foot string
}

func NewJSONP(ctx *types.HttpContext) *jsonp {
	j := &jsonp{}
	return j.New(ctx)
}

func (j *jsonp) New(ctx *types.HttpContext) *jsonp {
	j.polling = NewPolling(ctx)

	j.head = "___eio[" + regexp.MustCompile(`[^0-9]`).ReplaceAllString(ctx.Query().Peek("j"), "") + "]("
	j.foot = ");"

	j.onData = j.JSONPOnData
	j.doWrite = j.JSONPDoWrite
	return j
}

// Handles incoming data.
// Due to a bug in \n handling by browsers, we expect a escaped string.
func (j *jsonp) JSONPOnData(data types.BufferInterface) {
	if data, err := url.ParseQuery(data.String()); err == nil {
		if data.Has("d") {
			r := regexp.MustCompile(rSlashes)
			_data := r.ReplaceAllStringFunc(data.Get("d"), func(m string) string {
				if parts := r.FindStringSubmatch(m); parts[1] != "" {
					return parts[0]
				}
				return "\n"
			})
			// client will send already escaped newlines as \\\\n and newlines as \\n
			// \\n must be replaced with \n and \\\\n with \\n
			j.PollingOnData(types.NewStringBufferString(regexp.MustCompile(rDoubleSlashes).ReplaceAllString(_data, "\\n")))
		}
	} else {
		utils.Log().Debug(`jsonp OnData error "%v"`, err)
	}
}

func (j *jsonp) JSONPDoWrite(data types.BufferInterface, options *packet.Options, callback types.Callable) {
	// prepare response
	res := types.NewStringBufferString(j.head)
	encoder := json.NewEncoder(res)
	// we must output valid javascript, not valid json
	// see: http://timelessrepo.com/json-isnt-a-javascript-subset
	if err := encoder.Encode(data.String()); err == nil {
		// Since 1.18 the following source code is very annoying '\n' bytes
		res.Truncate(res.Len() - 1) // '\n' 😑
		res.WriteString(j.foot)
		j.PollingDoWrite(res, options, callback)
	} else {
		utils.Log().Debug(`jsonp DoWrite error "%v"`, err)
	}
}