package common

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/andybalholm/brotli"
	"golang.org/x/net/proxy"
)

var (
	BING_CHAT_DOMAIN = "https://sydney.bing.com"
	BING_CHAT_URL, _ = url.Parse(BING_CHAT_DOMAIN + "/sydney/ChatHub")
	BING_URL, _      = url.Parse("https://www.bing.com")
	// EDGE_SVC_URL, _     = url.Parse("https://edgeservices.bing.com")
	KEEP_REQ_HEADER_MAP = map[string]bool{
		"Accept":                   true,
		"Accept-Encoding":          true,
		"Accept-Language":          true,
		"Referer":                  true,
		"Connection":               true,
		"Cookie":                   true,
		"Upgrade":                  true,
		"User-Agent":               true,
		"Sec-Websocket-Extensions": true,
		"Sec-Websocket-Key":        true,
		"Sec-Websocket-Version":    true,
		"X-Request-Id":             true,
		"X-Forwarded-For":          true,
	}
	DEL_LOCATION_DOMAINS = []string{
		"https://cn.bing.com",
		"https://www.bing.com",
	}
	USER_TOKEN_COOKIE_NAME     = "_U"
	RAND_IP_COOKIE_NAME        = "BingAI_Rand_IP"
	PROXY_WEB_PREFIX_PATH      = "/web/"
	PROXY_WEB_PAGE_PATH        = PROXY_WEB_PREFIX_PATH + "index.html"
	USER_TOKEN_ENV_NAME_PREFIX = "Go_Proxy_BingAI_USER_TOKEN"
	USER_TOKEN_LIST            []string
	RAND_COOKIE_INDEX_NAME     = "BingAI_Rand_CK"
)

func init() {
	initUserToken()
}

func initUserToken() {
	for _, env := range os.Environ() {
		if strings.HasPrefix(env, USER_TOKEN_ENV_NAME_PREFIX) {
			parts := strings.SplitN(env, "=", 2)
			USER_TOKEN_LIST = append(USER_TOKEN_LIST, parts[1])
		}
	}
}

func NewSingleHostReverseProxy(target *url.URL) *httputil.ReverseProxy {
	originalScheme := "http"
	httpsSchemeName := "https"
	var originalHost string
	var originalPath string
	var originalDomain string
	var randIP string
	var resCKRandIndex string
	director := func(req *http.Request) {
		if req.URL.Scheme == httpsSchemeName || req.Header.Get("X-Forwarded-Proto") == httpsSchemeName {
			originalScheme = httpsSchemeName
		}
		originalHost = req.Host
		originalPath = req.URL.Path
		originalDomain = fmt.Sprintf("%s:%s", originalScheme, originalHost)

		req.URL.Scheme = target.Scheme
		req.URL.Host = target.Host
		req.Host = target.Host

		originalRefer := req.Referer()
		if originalRefer != "" && !strings.Contains(originalRefer, PROXY_WEB_PAGE_PATH) {
			req.Header.Set("Referer", strings.ReplaceAll(originalRefer, originalDomain, BING_URL.String()))
		} else {
			req.Header.Set("Referer", fmt.Sprintf("%s/search?q=Bing+AI", BING_URL.String()))
		}

		// 同一会话尽量保持相同的随机IP
		ckRandIP, _ := req.Cookie(RAND_IP_COOKIE_NAME)
		if ckRandIP != nil && ckRandIP.Value != "" {
			randIP = ckRandIP.Value
		}
		if randIP == "" {
			randIP = GetRandomIP()
		}
		req.Header.Set("X-Forwarded-For", randIP)

		// 未登录用户
		ckUserToken, _ := req.Cookie(USER_TOKEN_COOKIE_NAME)
		if ckUserToken == nil || ckUserToken.Value == "" {
			randCKIndex, randCkVal := getRandCookie(req)
			if randCkVal != "" {
				resCKRandIndex = strconv.Itoa(randCKIndex)
				req.AddCookie(&http.Cookie{
					Name:  USER_TOKEN_COOKIE_NAME,
					Value: randCkVal,
				})
			}
			// ua := req.UserAgent()
			// if !strings.Contains(ua, "iPhone") || !strings.Contains(ua, "Mobile") {
			// 	req.Header.Set("User-Agent", "iPhone Mobile "+ua)
			// }
		}
		req.Header.Set("User-Agent", "Mozilla/5.0 (iPhone; CPU iPhone OS 15_7 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/15.7 Mobile/15E148 Safari/605.1.15 BingSapphire/1.0.410427012")

		for hKey, _ := range req.Header {
			if _, ok := KEEP_REQ_HEADER_MAP[hKey]; !ok {
				req.Header.Del(hKey)
			}
		}

		// reqHeaderByte, _ := json.Marshal(req.Header)
		// log.Println("剩余请求头 ： ", string(reqHeaderByte))
	}
	//改写返回信息
	modifyFunc := func(res *http.Response) error {
		contentType := res.Header.Get("Content-Type")
		if strings.Contains(contentType, "text/javascript") {
			contentEncoding := res.Header.Get("Content-Encoding")
			switch contentEncoding {
			case "gzip":
				// log.Println("ContentEncoding : ", contentEncoding, " Path : ", originalPath)
				modifyGzipBody(res, originalScheme, originalHost)
			case "br":
				// log.Println("ContentEncoding : ", contentEncoding, " Path : ", originalPath)
				modifyBrBody(res, originalScheme, originalHost)
			default:
				log.Println("ContentEncoding default : ", contentEncoding, " Path : ", originalPath)
				modifyDefaultBody(res, originalScheme, originalHost)
			}
		}

		// 修改响应 cookie 域
		// resCookies := res.Header.Values("Set-Cookie")
		// if len(resCookies) > 0 {
		// 	for i, v := range resCookies {
		// 		resCookies[i] = strings.ReplaceAll(strings.ReplaceAll(v, ".bing.com", originalHost), "bing.com", originalHost)
		// 	}
		// }

		// 设置随机ip cookie
		ckRandIP := &http.Cookie{
			Name:  RAND_IP_COOKIE_NAME,
			Value: randIP,
			Path:  "/",
		}
		res.Header.Set("Set-Cookie", ckRandIP.String())

		// 设置服务器 cookie 对应索引
		if resCKRandIndex != "" {
			ckRandIndex := &http.Cookie{
				Name:  RAND_COOKIE_INDEX_NAME,
				Value: resCKRandIndex,
				Path:  "/",
			}
			res.Header.Set("Set-Cookie", ckRandIndex.String())
		}

		// 删除 CSP
		res.Header.Del("Content-Security-Policy-Report-Only")
		res.Header.Del("Report-To")

		// 删除重定向前缀域名 cn.bing.com www.bing.com 等
		location := res.Header.Get("Location")
		if location != "" {
			for _, delLocationDomain := range DEL_LOCATION_DOMAINS {
				if strings.HasPrefix(location, delLocationDomain) {
					res.Header.Set("Location", location[len(delLocationDomain):])
					log.Println("Del Location Domain ：", location)
				}
			}
		}

		// 跨域
		res.Header.Set("Access-Control-Allow-Origin", "*")
		res.Header.Set("Access-Control-Allow-Methods", "*")
		res.Header.Set("Access-Control-Allow-Headers", "*")

		return nil
	}
	errorHandler := func(res http.ResponseWriter, req *http.Request, err error) {
		log.Println("代理异常 ：", err)
		res.Write([]byte(err.Error()))
	}

	// tr := &http.Transport{
	// 	TLSClientConfig: &tls.Config{
	// 		// 如果只设置 InsecureSkipVerify: true对于这个问题不会有任何改变
	// 		InsecureSkipVerify: true,
	// 		ClientAuth:         tls.NoClientCert,
	// 	},
	// }

	// 代理请求   请求回来的内容   报错自动调用
	reverseProxy := &httputil.ReverseProxy{
		Director:       director,
		ModifyResponse: modifyFunc,
		ErrorHandler:   errorHandler,
	}

	// socks
	socksUrl := os.Getenv("Go_Proxy_BingAI_SOCKS_URL")
	if socksUrl != "" {
		socksUser := os.Getenv("Go_Proxy_BingAI_SOCKS_USER")
		socksPwd := os.Getenv("Go_Proxy_BingAI_SOCKS_PWD")
		var socksAuth *proxy.Auth
		if socksUser != "" && socksPwd != "" {
			socksAuth = &proxy.Auth{
				User:     socksUser,
				Password: socksPwd,
			}
		}
		s5Proxy, err := proxy.SOCKS5("tcp", socksUrl, socksAuth, proxy.Direct)
		if err != nil {
			panic(err)
		}
		tr := &http.Transport{
			Dial: s5Proxy.Dial,
		}
		reverseProxy.Transport = tr
	}

	return reverseProxy
}

// return cookie index and cookie
func getRandCookie(req *http.Request) (int, string) {
	utLen := len(USER_TOKEN_LIST)
	if utLen == 0 {
		return 0, ""
	}
	if utLen == 1 {
		return 0, USER_TOKEN_LIST[0]
	}

	ckRandIndex, _ := req.Cookie(RAND_COOKIE_INDEX_NAME)
	if ckRandIndex != nil && ckRandIndex.Value != "" {
		tmpIndex, err := strconv.Atoi(ckRandIndex.Value)
		if err != nil {
			log.Println("ckRandIndex err ：", err)
			return 0, ""
		}
		if tmpIndex < utLen {
			return tmpIndex, USER_TOKEN_LIST[tmpIndex]
		}
	}
	seed := time.Now().UnixNano()
	rng := rand.New(rand.NewSource(seed))
	randIndex := rng.Intn(len(USER_TOKEN_LIST))
	return randIndex, USER_TOKEN_LIST[randIndex]
}

func replaceResBody(originalBody string, originalScheme string, originalHost string) string {
	modifiedBodyStr := originalBody
	originalDomain := fmt.Sprintf("%s://%s", originalScheme, originalHost)

	if strings.Contains(modifiedBodyStr, BING_URL.String()) {
		modifiedBodyStr = strings.ReplaceAll(modifiedBodyStr, BING_URL.String(), originalDomain)
	}

	// 对话暂时支持国内网络，而且 Vercel 还不支持 Websocket ，先不用
	// if strings.Contains(modifiedBodyStr, BING_CHAT_DOMAIN) {
	// 	modifiedBodyStr = strings.ReplaceAll(modifiedBodyStr, BING_CHAT_DOMAIN, originalDomain)
	// }

	// if strings.Contains(modifiedBodyStr, "https://www.bingapis.com") {
	// 	modifiedBodyStr = strings.ReplaceAll(modifiedBodyStr, "https://www.bingapis.com", "https://bing.vcanbb.top")
	// }
	return modifiedBodyStr
}

func modifyGzipBody(res *http.Response, originalScheme string, originalHost string) error {
	gz, err := gzip.NewReader(res.Body)
	if err != nil {
		return err
	}
	defer gz.Close()

	bodyByte, err := io.ReadAll(gz)
	if err != nil {
		return err
	}
	originalBody := string(bodyByte)
	modifiedBodyStr := replaceResBody(originalBody, originalScheme, originalHost)
	// 修改响应内容
	modifiedBody := []byte(modifiedBodyStr)
	// gzip 压缩
	var buf bytes.Buffer
	writer := gzip.NewWriter(&buf)
	defer writer.Close()

	_, err = writer.Write(modifiedBody)
	if err != nil {
		return err
	}
	err = writer.Flush()
	if err != nil {
		return err
	}
	err = writer.Close()
	if err != nil {
		return err
	}

	// 修改 Content-Length 头
	res.Header.Set("Content-Length", strconv.Itoa(buf.Len()))
	// 修改响应内容
	res.Body = io.NopCloser(&buf)

	return nil
}

func modifyBrBody(res *http.Response, originalScheme string, originalHost string) error {
	reader := brotli.NewReader(res.Body)
	var uncompressed bytes.Buffer
	uncompressed.ReadFrom(reader)

	originalBody := uncompressed.String()

	modifiedBodyStr := replaceResBody(originalBody, originalScheme, originalHost)

	// 修改响应内容
	modifiedBody := []byte(modifiedBodyStr)
	// br 压缩
	var buf bytes.Buffer
	writer := brotli.NewWriter(&buf)
	writer.Write(modifiedBody)
	writer.Close()

	// 修改 Content-Length 头
	// res.ContentLength = int64(buf.Len())
	res.Header.Set("Content-Length", strconv.Itoa(buf.Len()))
	// 修改响应内容
	res.Body = io.NopCloser(bytes.NewReader(buf.Bytes()))

	return nil
}

func modifyDefaultBody(res *http.Response, originalScheme string, originalHost string) error {
	bodyByte, err := io.ReadAll(res.Body)
	if err != nil {
		return err
	}
	originalBody := string(bodyByte)
	modifiedBodyStr := replaceResBody(originalBody, originalScheme, originalHost)
	// 修改响应内容
	modifiedBody := []byte(modifiedBodyStr)

	// 修改 Content-Length 头
	// res.ContentLength = int64(buf.Len())
	res.Header.Set("Content-Length", strconv.Itoa(len(modifiedBody)))
	// 修改响应内容
	res.Body = io.NopCloser(bytes.NewReader(modifiedBody))

	return nil
}
