package api

import (
	"bytes"
	"io"
	"io/ioutil"
	"net/http"
	"regexp"
	"strconv"
	"unicode"

	"github.com/octanolabs/go-spectrum/models"

	"github.com/gin-gonic/gin"

	json "github.com/json-iterator/go"
	"github.com/ubiq/go-ubiq/log"
)

// v2 helper functions
// TODO: add 404 not found response for requests that don't match

var legacyHandlers = map[*regexp.Regexp]func(re *regexp.Regexp, url string) (io.Reader, int64, string){

	regexp.MustCompile(`^/v2/(?:status)$`): jsonhttphelper("explorer_status"),

	regexp.MustCompile(`^/v2/(?:latest)$`):                        jsonhttphelper("explorer_latestBlock"),
	regexp.MustCompile(`^/v2/(?:latestblocks)/(?P<params>(.*))$`): jsonhttphelper("explorer_latestBlocks"),
	regexp.MustCompile(`^/v2/(?:blockbyhash)/(?P<params>(.*))$`):  jsonhttphelper("explorer_blockByHash"),
	regexp.MustCompile(`^/v2/(?:block)/(?P<params>([^/]*))$`):     jsonhttphelper("explorer_blockByNumber"),
	regexp.MustCompile(`^/v2/(?:block)/(?P<params>(.*))/txns$`):   jsonhttphelper("explorer_transactionsByBlockNumber"),

	regexp.MustCompile(`^/v2/(?:latestuncles)/(?P<params>(.*))$`): jsonhttphelper("explorer_latestUncles"),
	regexp.MustCompile(`^/v2/(?:uncle)/(?P<params>(.*))$`):        jsonhttphelper("explorer_uncleByHash"),

	regexp.MustCompile(`^/v2/(?:forkedblock)/(?P<params>(.*))$`):        jsonhttphelper("explorer_forkedBlockByNumber"),
	regexp.MustCompile(`^/v2/(?:latestforkedblocks)/(?P<params>(.*))$`): jsonhttphelper("explorer_latestForkedBlocks"),

	regexp.MustCompile(`^/v2/(?:latesttransactions)/(?P<params>(.*))$`):    jsonhttphelper("explorer_latestTransactions"),
	regexp.MustCompile(`^/v2/(?:transaction)/(?P<params>(.*))$`):           jsonhttphelper("explorer_transactionByHash"),
	regexp.MustCompile(`^/v2/(?:latestaccounttxns)/(?P<params>(.*))$`):     jsonhttphelper("explorer_latestTransactionsByAccount"),
	regexp.MustCompile(`^/v2/(?:transactionbycontract)/(?P<params>(.*))$`): jsonhttphelper("explorer_transactionByContractAddress"),

	regexp.MustCompile(`^/v2/(?:latesttokentransfers)/(?P<params>(.*))$`):         jsonhttphelper("explorer_latestTokenTransfers"),
	regexp.MustCompile(`^/v2/(?:latestaccounttokentxns)/(?P<params>(.*))$`):       jsonhttphelper("explorer_latestTokenTransfersByAccount"),
	regexp.MustCompile(`^/v2/(?:tokentransfersbyaccount)/(?P<params>(.*)/(.*))$`): jsonhttphelper("explorer_tokenTransfersByAccount"),
	regexp.MustCompile(`^/v2/(?:latesttransfersbytoken)/(?P<params>(.*))$`):       jsonhttphelper("explorer_latestTransfersOfToken"),

	//regexp.MustCompile(`^/(?:charts)/(?P<params>(?P<chart>.)/(?P<limit>.))$`): jsonhttphelper("explorer_"),
	//regexp.MustCompile(`^/(?:supply)/(?P<params>(?P<symbol>.))$`):             jsonhttphelper("explorer_"),
	//regexp.MustCompile(`^/(?:(geodata))$`):                                      jsonhttphelper("explorer_"),
}

func jsonhttphelper(method string) func(*regexp.Regexp, string) (io.Reader, int64, string) {
	return func(re *regexp.Regexp, url string) (io.Reader, int64, string) {
		var (
			expanded []byte
			result   []byte
			fields   [][]byte
			req      struct {
				Id      int           `json:"id"`
				JsonRpc string        `json:"json_rpc"`
				Method  string        `json:"method"`
				Params  []interface{} `json:"params"`
			}
		)

		template := "${params}"

		// For each match of the regex in the content.
		for _, submatches := range re.FindAllStringSubmatchIndex(url, -1) {
			// Apply the captured submatches to the template and append the output
			// to the result.
			expanded = re.ExpandString(expanded, template, url, submatches)
		}

		fields = bytes.FieldsFunc(expanded, func(c rune) bool { return !unicode.IsLetter(c) && !unicode.IsNumber(c) })

		req.Id = 88
		req.JsonRpc = "2.0"
		req.Method = method
		req.Params = make([]interface{}, len(fields))

		for k, v := range fields {
			num, err := strconv.ParseInt(string(v), 10, 64)
			if err != nil {
				if _, ok := err.(*strconv.NumError); ok {
					log.Debug("Param is not number, setting as string", "method", method, "value", string(v))
					req.Params[k] = string(v)
				} else {
					log.Debug("Unexpected error converting param, setting as string", "method", method, "value", string(v), "err", err)
					req.Params[k] = string(v)
				}
			} else {
				req.Params[k] = num
			}
		}

		result, err := json.Marshal(req)
		if err != nil {
			log.Error("Error: couldn't parse regex", "err", err)
		}

		return io.LimitReader(bytes.NewReader(result), int64(len(result))), int64(len(result)), method
	}
}

func ConvertJSONHTTPReq(r *http.Request) (io.ReadCloser, int64, string) {

	var (
		res    io.Reader
		length int64
		method string
	)

	for k, handler := range legacyHandlers {
		if k.MatchString(r.URL.Path) {
			res, length, method = handler(k, r.URL.Path)
		}
	}

	return ioutil.NopCloser(res), length, method
}

func ParseJsonRequest(r *http.Request) (string, []json.RawMessage, io.ReadCloser) {

	var (
		b   = make([]byte, r.ContentLength)
		req struct {
			Method string            `json:"method"`
			Params []json.RawMessage `json:"params"`
		}
	)

	b, err := ioutil.ReadAll(io.LimitReader(r.Body, r.ContentLength))
	if err != nil {
		log.Error("Util: couldn't write request body to buffer", "err", err)
	}

	err = json.Unmarshal(b, &req)

	if err != nil {
		log.Error("Error: couldn't unmarshal body", "err", err)
	}

	return req.Method, req.Params, ioutil.NopCloser(bytes.NewReader(b))

}

// v2 handlers

// explorer_latestBlocks explorer_latestUncles explorer_latestTransactions explorer_latestTokenTransfers
// for these 4 methods the legacy api returned the totals for these collections, and we append that info
// directly into response via v2ConvertResponseWriter

func v2ConvertRequest() gin.HandlerFunc {
	return func(context *gin.Context) {
		newReader, length, method := ConvertJSONHTTPReq(context.Request)

		l := strconv.FormatInt(length, 10)

		context.Request.Body = newReader
		context.Request.ContentLength = length
		context.Request.Header.Set("Content-Length", l)
		context.Request.Header.Set("Content-Type", "application/json")

		context.Set("method", method)

	}
}

func v2ConvertResponse(handlers v3api) gin.HandlerFunc {
	return func(context *gin.Context) {

		var p string

		param, ok := context.Get("method")

		if ok {
			p, _ = param.(string)
		}

		context.Writer = v2ConvertResponseWriter{ResponseWriter: context.Writer, statusHandler: handlers, method: p}
	}
}

type v2ConvertResponseWriter struct {
	statusHandler interface{ Status() (models.Store, error) }
	method        string
	gin.ResponseWriter
}

func (r v2ConvertResponseWriter) Write(b []byte) (int, error) {

	var (
		req struct {
			Body json.RawMessage `json:"result"`
		}
	)

	err := json.Unmarshal(b, &req)

	if err != nil {
		log.Error("Error: couldn't marshal response body", "err", err)
	}

	if string(req.Body) == "null" {
		req.Body = []byte("[]")
	}

	store, err := r.statusHandler.Status()

	if err != nil {
		log.Error("error with responsewriter", "err", err)
		return r.ResponseWriter.Write(req.Body)
	}

	switch r.method {
	case "explorer_latestBlocks":
		msg, err := json.Marshal(map[string]interface{}{"blocks": req.Body, "total": store.LatestBlock.Number + 1})
		if err != nil {
			break
		}
		return r.ResponseWriter.Write(msg)
	case "explorer_latestUncles":
		msg, err := json.Marshal(map[string]interface{}{"uncles": req.Body, "total": store.TotalUncles})
		if err != nil {
			break
		}
		return r.ResponseWriter.Write(msg)
	case "explorer_latestTransactions":
		msg, err := json.Marshal(map[string]interface{}{"txns": req.Body, "total": store.TotalTransactions})
		if err != nil {
			break
		}
		return r.ResponseWriter.Write(msg)
	case "explorer_latestTokenTransfers":
		msg, err := json.Marshal(map[string]interface{}{"txns": req.Body, "total": store.TotalTokenTransfers})
		if err != nil {
			break
		}
		return r.ResponseWriter.Write(msg)
	}

	return r.ResponseWriter.Write(req.Body)
}
