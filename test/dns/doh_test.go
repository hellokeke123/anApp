package test

import (
	"encoding/base64"
	"fmt"
	"github.com/miekg/dns"
	"io/ioutil"
	"net/http"
	"os"
	"testing"
)

func TestDOH(t *testing.T) {
	query := dns.Msg{}
	query.SetQuestion("www,baidu.com.", dns.TypeA)
	msg, _ := query.Pack()
	b64 := base64.RawURLEncoding.EncodeToString(msg)
	resp, err := http.Get("https://223.5.5.5/dns-query?dns=" + b64)
	if err != nil {
		fmt.Printf("Send query error, err:%v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()
	bodyBytes, _ := ioutil.ReadAll(resp.Body)
	response := dns.Msg{}
	response.Unpack(bodyBytes)
	fmt.Printf("Dns answer is :%v\n", response.String())
}
