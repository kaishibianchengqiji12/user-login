package accesslog

import (
	"encoding/json"
	"github.com/kataras/iris/v12/core/memstore"
	"github.com/kataras/iris/v12/middleware/accesslog"
	"github.com/mailru/easyjson/jwriter"
	"io"
	"net/url"
)

// UniformJSON is a Formatter type for JSON logs.
type UniformJSON struct {
	EscapeHTML bool
	HumanTime  bool
	//服务标识字段
	ServerName  string
	Environment string
	InstanceKey string

	ac *accesslog.AccessLog
}

// SetOutput creates the json encoder writes to the "dest".
// It's called automatically by the middleware when this Formatter is used.
func (f *UniformJSON) SetOutput(dest io.Writer) {
	f.ac, _ = dest.(*accesslog.AccessLog)
}

var (
	timestampKeyB       = []byte(`"timestamp":`)
	timestampKeyIndentB = append(timestampKeyB, ' ')
)

const (
	newLine = '\n'
)

// Format prints the logs in JSON format.
// Writes to the destination directly,
// locks on each Format call.
func (f *UniformJSON) Format(log *accesslog.Log) (bool, error) {
	err := f.writeEasyJSON(log)
	return true, err
}

func (f *UniformJSON) writeEasyJSON(in *accesslog.Log) error {
	out := &jwriter.Writer{NoEscapeHTML: !f.EscapeHTML}

	out.RawByte('{')
	first := true
	_ = first
	{
		const prefix string = ",\"timestamp\":"
		if first {
			first = false
			out.RawString(prefix[1:])
		} else {
			out.RawString(prefix)
		}

		if f.HumanTime {
			t := in.Now.Format(in.TimeFormat)
			out.String(t)
		} else {
			out.Int64(in.Timestamp)
		}
	}
	{
		const prefix string = ",\"server_name\":"
		out.RawString(prefix)
		out.String(f.ServerName)
	}
	{
		const prefix string = ",\"environment\":"
		out.RawString(prefix)
		out.String(f.Environment)
	}
	{
		const prefix string = ",\"instance_key\":"
		out.RawString(prefix)
		out.String(f.InstanceKey)
	}
	{
		const prefix string = ",\"log_type\":"
		out.RawString(prefix)
		out.String("access")
	}
	{
		const prefix string = ",\"latency\":"
		out.RawString(prefix)
		//改动点：延迟保留到ms级别
		out.Int64(int64(in.Latency) / 1000000)
	}
	{
		const prefix string = ",\"code\":"
		out.RawString(prefix)
		out.Int(int(in.Code))
	}
	{
		const prefix string = ",\"method\":"
		out.RawString(prefix)
		out.String(in.Method)
	}
	{
		const prefix string = ",\"path\":"
		out.RawString(prefix)
		out.String(in.Path)
	}
	if in.IP != "" {
		const prefix string = ",\"ip\":"
		out.RawString(prefix)
		out.String(in.IP)
	}
	if len(in.Query) != 0 {
		const prefix string = ",\"query\":"
		out.RawString(prefix)
		{
			out.RawByte('[')
			for v4, v5 := range in.Query {
				if v4 > 0 {
					out.RawByte(',')
				}
				easyJSONStringEntry(out, v5)
			}
			out.RawByte(']')
		}
	}
	if len(in.PathParams) != 0 {
		const prefix string = ",\"params\":"
		out.RawString(prefix)
		{
			out.RawByte('[')
			for v6, v7 := range in.PathParams {
				if v6 > 0 {
					out.RawByte(',')
				}
				easyJSONEntry(out, v7)
			}
			out.RawByte(']')
		}
	}
	if len(in.Fields) != 0 {
		const prefix string = ",\"fields\":"
		out.RawString(prefix)
		{
			var reqidEntry memstore.Entry
			var usernameEntry memstore.Entry

			out.RawByte('[')
			for v8, v9 := range in.Fields {
				// v8 是index
				if v8 > 0 {
					out.RawByte(',')
				}
				if v9.Key == "url_params" {
					easyUrlEncodedEntry(out, v9)
				} else {
					easyJSONEntry(out, v9)
				}

				if v9.Key == "req_id" {
					reqidEntry = v9
					continue
				}

				if v9.Key == "user" {
					usernameEntry = v9
					continue
				}

			}
			out.RawByte(']')

			const reqidPrefix string = ",\"req_id\":"
			out.RawString(reqidPrefix)
			out.Raw(json.Marshal(reqidEntry.ValueRaw))

			const userPrefix string = ",\"user\":"
			out.RawString(userPrefix)
			out.Raw(json.Marshal(usernameEntry.ValueRaw))
		}
	}
	if in.Logger.RequestBody {
		const prefix string = ",\"request\":"
		out.RawString(prefix)
		out.String(string(in.Request))
	}
	if in.Logger.ResponseBody {

		const prefix string = ",\"response\":"
		out.RawString(prefix)
		out.String(string(in.Response))

	}
	if in.BytesReceived != 0 {
		const prefix string = ",\"bytes_received\":"
		out.RawString(prefix)
		out.Int(int(in.BytesReceived))
	}
	if in.BytesSent != 0 {
		const prefix string = ",\"bytes_sent\":"
		out.RawString(prefix)
		out.Int(int(in.BytesSent))
	}
	out.RawByte('}')
	out.RawByte(newLine)

	if out.Error != nil {
		return out.Error
	}
	f.ac.Write(out.Buffer.BuildBytes())
	return nil
}

func easyUrlEncodedEntry(out *jwriter.Writer, in memstore.Entry) {
	out.RawByte('{')
	first := true
	_ = first
	{
		const prefix string = ",\"key\":"
		out.RawString(prefix[1:])
		out.String(string(in.Key))
	}
	{
		const prefix string = ",\"value\":"
		out.RawString(prefix)

		valueRaw, _ := in.ValueRaw.(string)
		decodeUrl, err := url.QueryUnescape(valueRaw)
		if err != nil {
			out.Error = err
			return
		}
		out.String(decodeUrl)
	}
	out.RawByte('}')
}

func easyJSONEntry(out *jwriter.Writer, in memstore.Entry) {
	out.RawByte('{')
	first := true
	_ = first
	{
		const prefix string = ",\"key\":"
		out.RawString(prefix[1:])
		out.String(string(in.Key))
	}
	{
		const prefix string = ",\"value\":"
		out.RawString(prefix)
		if m, ok := in.ValueRaw.(json.Marshaler); ok {
			out.Raw(m.MarshalJSON())
		} else {
			out.Raw(json.Marshal(in.ValueRaw))
		}
	}
	out.RawByte('}')
}

func easyJSONStringEntry(out *jwriter.Writer, in memstore.StringEntry) {
	out.RawByte('{')
	first := true
	_ = first
	{
		const prefix string = ",\"key\":"
		out.RawString(prefix[1:])
		out.String(string(in.Key))
	}
	{
		const prefix string = ",\"value\":"
		out.RawString(prefix)
		out.String(string(in.Value))
	}
	out.RawByte('}')
}
