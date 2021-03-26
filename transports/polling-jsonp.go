package transports

type JSONP interface {
	Polling
}

type jsonp struct {
	*polling
}

func NewJSONP(res interface{}) JSONP {
	j := &JSONP{}
	// this.head = "___eio[" + (req._query.j || "").replace(/[^0-9]/g, "") + "](";
	// this.foot = ");";

	return j
}

/**
 * Handles incoming data.
 * Due to a bug in \n handling by browsers, we expect a escaped string.
 *
 * @api public
 */
func (j *JSONP) OnData(data io.Reader) {
	// we leverage the qs module so that we get built-in DoS protection
	// and the fast alternative to decodeURIComponent
	// data = qs.parse(data).d;
	// if ("string" === typeof data) {
	//   // client will send already escaped newlines as \\\\n and newlines as \\n
	//   // \\n must be replaced with \n and \\\\n with \\n
	//   data = data.replace(rSlashes, function(match, slashes) {
	//     return slashes ? match : "\n";
	//   });
	//   super.onData(data.replace(rDoubleSlashes, "\\n"));
	// }
}

func (j *JSONP) DoWrite(data, options, callback) {
	// we must output valid javascript, not valid json
	// // see: http://timelessrepo.com/json-isnt-a-javascript-subset
	// const js = JSON.stringify(data)
	//   .replace(/\u2028/g, "\\u2028")
	//   .replace(/\u2029/g, "\\u2029");

	// // prepare response
	// data = this.head + js + this.foot;

	// super.doWrite(data, options, callback);
}
