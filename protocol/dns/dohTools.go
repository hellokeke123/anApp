package dns

import (
	"bytes"
	"encoding/base64"
	"errors"
	mdns "github.com/miekg/dns"
	"log"
	"net/http"
)

func SendDoh(dohUrl string, source *bytes.Buffer) (buff *bytes.Buffer, err error) {
	defer func() {
		err := recover()
		if err != nil {
			log.Println("dns-doh错误发送：", err)
		}
		return
	}()

	if source == nil || source.Len() <= 0 {
		log.Println("doh:", "数据错误")
		return nil, errors.New("doh:来源数据错误")
	} else {
		msg := mdns.Msg{}
		msg.Unpack(source.Bytes())
		log.Println("doh:", msg.Question)
	}

	b64 := base64.RawURLEncoding.EncodeToString(source.Bytes())
	resp, err := http.Get(dohUrl + "?dns=" + b64)
	if err != nil {
		return nil, err
	} else {
		defer resp.Body.Close()
		buff := &bytes.Buffer{}
		_, err := buff.ReadFrom(resp.Body)

		return buff, err

	}

}
