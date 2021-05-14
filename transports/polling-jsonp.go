package transports

import (
	"encoding/json"
	"github.com/zishang520/engine.io/types"
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
	j := &jsonp{
		polling: NewPolling(ctx),

		head: "___eio[" + regexp.MustCompile(`[^0-9]`).ReplaceAllString(string(ctx.QueryArgs().Peek("j")), "") + "](",
		foot: ");",
	}
	return j
}

/**
 * Handles incoming data.
 * Due to a bug in \n handling by browsers, we expect a escaped string.
 *
 * @api public
 */
func (j *jsonp) OnData(data io.Reader) {
	// we leverage the qs module so that we get built-in DoS protection
	// and the fast alternative to decodeURIComponent
	u, err := url.ParseQuery(data)
	if err != nil {
		return
	}
	if _, ok := u["d"]; !ok {
		return
	}
	r := regexp.MustCompile(rSlashes)
	_data := r.ReplaceAllStringFunc(u.Get("d"), func(m string) string {
		if parts := r.FindStringSubmatch(m); parts[1] != "" {
			return parts[0]
		}
		return "\n"
	})
	// client will send already escaped newlines as \\\\n and newlines as \\n
	// \\n must be replaced with \n and \\\\n with \\n
	j.polling.OnData(regexp.MustCompile(rDoubleSlashes).ReplaceAllString(_data, "\\n"))
}

func (j *jsonp) DoWrite(data io.Reader, options *packet.Option, callback types.Fn) {
	if c, ok := data.(io.Closer); ok {
		defer c.Close()
	}
	// we must output valid javascript, not valid json
	// see: http://timelessrepo.com/json-isnt-a-javascript-subset
	// const js = JSON.stringify(data)
	//   .replace(/\u2028/g, "\\u2028")
	//   .replace(/\u2029/g, "\\u2029");

	// prepare response
	res := types.NewStringBufferString(j.head)
	json.NewEncoder(res).Encode(data) // 有问题
	res.WriteString(j.foot)

	j.polling.DoWrite(res, options, callback)
}
