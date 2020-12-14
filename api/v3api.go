package api

import (
	"bytes"
	"github.com/octanolabs/go-spectrum/models"
	"io"
	"io/ioutil"
	"net/http"
	"regexp"
	"strconv"
	"unicode"

	"github.com/gin-gonic/gin"

	json "github.com/json-iterator/go"
	"github.com/ubiq/go-ubiq/v3/log"
)

// v3 helper functions
// TODO: add 404 not found response for requests that don't match

var legacyHandlers = map[*regexp.Regexp]func(re *regexp.Regexp, url string) (io.Reader, int64, string){

	//TODO: add back ?supplyOnly query for coinmarketcap
	regexp.MustCompile(`^/v3/(?:status)$`): jsonhttphelper("explorer_status"),

	regexp.MustCompile(`^/v3/(?:latest)$`):                        jsonhttphelper("explorer_latestBlock"),
	regexp.MustCompile(`^/v3/(?:latestblocks)/(?P<params>(.*))$`): jsonhttphelper("explorer_latestBlocks"),
	regexp.MustCompile(`^/v3/(?:blockbyhash)/(?P<params>(.*))$`):  jsonhttphelper("explorer_blockByHash"),
	regexp.MustCompile(`^/v3/(?:block)/(?P<params>([^/]*))$`):     jsonhttphelper("explorer_blockByNumber"),
	regexp.MustCompile(`^/v3/(?:block)/(?P<params>(.*))/txns$`):   jsonhttphelper("explorer_transactionsByBlockNumber"),

	regexp.MustCompile(`^/v3/(?:latestuncles)/(?P<params>(.*))$`): jsonhttphelper("explorer_latestUncles"),
	regexp.MustCompile(`^/v3/(?:uncle)/(?P<params>(.*))$`):        jsonhttphelper("explorer_uncleByHash"),

	regexp.MustCompile(`^/v3/(?:forkedblock)/(?P<params>(.*))$`):        jsonhttphelper("explorer_forkedBlockByNumber"),
	regexp.MustCompile(`^/v3/(?:latestforkedblocks)/(?P<params>(.*))$`): jsonhttphelper("explorer_latestForkedBlocks"),

	regexp.MustCompile(`^/v3/(?:latesttransactions)/(?P<params>(.*))$`):    jsonhttphelper("explorer_latestTransactions"),
	regexp.MustCompile(`^/v3/(?:transaction)/(?P<params>(.*))$`):           jsonhttphelper("explorer_transactionByHash"),
	regexp.MustCompile(`^/v3/(?:latestaccounttxns)/(?P<params>(.*))$`):     jsonhttphelper("explorer_latestTransactionsByAccount"),
	regexp.MustCompile(`^/v3/(?:transactionbycontract)/(?P<params>(.*))$`): jsonhttphelper("explorer_transactionByContractAddress"),

	regexp.MustCompile(`^/v3/(?:latesttokentransfers)/(?P<params>(.*))$`):         jsonhttphelper("explorer_latestTokenTransfers"),
	regexp.MustCompile(`^/v3/(?:latestaccounttokentxns)/(?P<params>(.*))$`):       jsonhttphelper("explorer_latestTokenTransfersByAccount"),
	regexp.MustCompile(`^/v3/(?:tokentransfersbyaccount)/(?P<params>(.*)/(.*))$`): jsonhttphelper("explorer_tokenTransfersByAccount"),
	regexp.MustCompile(`^/v3/(?:latesttransfersbytoken)/(?P<params>(.*))$`):       jsonhttphelper("explorer_latestTransfersOfToken"),

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

func ParseJsonRequest(r *http.Request) (json.RawMessage, []json.RawMessage, io.ReadCloser) {

	var (
		b   = make([]byte, r.ContentLength)
		req struct {
			Method json.RawMessage   `json:"method"`
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

// v3 handlers

// explorer_latestBlocks explorer_latestUncles explorer_latestTransactions explorer_latestTokenTransfers
// for these 4 methods the legacy api returned the totals for these collections, and we append that info
// directly into response via v3ConvertResponseWriter

func v3ConvertRequest() gin.HandlerFunc {
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

func v3ConvertResponse() gin.HandlerFunc {
	return func(context *gin.Context) {

		_, ok := context.Request.URL.Query()["supplyOnly"]

		if context.Param("path") == "/status" && ok {
			context.Writer = v3SupplyOnlyResponseWriter{ResponseWriter: context.Writer}
		} else {
			context.Writer = v3ConvertResponseWriter{ResponseWriter: context.Writer}
		}
	}
}

type v3ConvertResponseWriter struct {
	gin.ResponseWriter
}

func (r v3ConvertResponseWriter) Write(b []byte) (int, error) {

	var (
		req struct {
			Body json.RawMessage `json:"result"`
		}
	)

	err := json.Unmarshal(b, &req)

	if err != nil {
		log.Error("Error: couldn't marshal response body", "err", err)
	}

	return r.ResponseWriter.Write(req.Body)
}

type v3SupplyOnlyResponseWriter struct {
	gin.ResponseWriter
}

func (r v3SupplyOnlyResponseWriter) Write(b []byte) (int, error) {

	var (
		req struct {
			Body models.Store `json:"result"`
		}
	)

	err := json.Unmarshal(b, &req)
	if err != nil {
		log.Error("Error: couldn't marshal response body", "err", err)
		return 0, err
	}

	bb, err := json.Marshal(req.Body.Supply)
	if err != nil {
		log.Error("Error: couldn't marshal response body", "err", err)
		return 0, err
	}

	return r.ResponseWriter.Write(bb)
}
