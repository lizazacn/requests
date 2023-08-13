package Requests

import (
	"crypto/tls"
	"encoding/json"
	"errors"
	"io"
	"io/ioutil"
	"net/http"
	URL "net/url"
	"reflect"
	"time"
)

type Response struct {
	http.Response
	Text []byte
	Json map[string]interface{}
	Map  map[string]interface{}
}

func Requests(method, url string, body io.Reader, header http.Header, format, skipHttpsVerify, proxy bool, proxyUrl *URL.URL) (*Response, error) {
	resp := new(Response)
	request, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, err
	}
	if header == nil {
		header = http.Header{}
		header.Add("User-Agent", "ELST Request/1.0 (Golang)")
	}
	request.Header = header

	client := &http.Client{}

	t := &http.Transport{
		MaxIdleConns:    10,
		MaxConnsPerHost: 10,
		TLSClientConfig: &tls.Config{InsecureSkipVerify: skipHttpsVerify},
	}
	client.Transport = t

	// 启用代理
	if proxy {
		t = &http.Transport{
			MaxIdleConns:    10,
			MaxConnsPerHost: 10,
			IdleConnTimeout: time.Duration(10) * time.Second,
			Proxy:           http.ProxyURL(proxyUrl),
			TLSClientConfig: &tls.Config{InsecureSkipVerify: skipHttpsVerify},
		}
		client.Transport = t
	}

	response, err := client.Do(request)
	if err != nil {
		return nil, err
	}
	CopyStruct(resp, response)

	if format {
		resp.Text, err = ioutil.ReadAll(response.Body)
		if err != nil {
			return resp, err
		}
		var respMap = make(map[string]interface{})
		err = json.Unmarshal(resp.Text, &respMap)
		if err != nil {
			return resp, errors.New("FORMAT DATA ERROR")
		}
		resp.Map = respMap
		resp.Json = respMap
	}
	return resp, nil
}

// CopyStruct
// dst 目标结构体，src 源结构体
// 必须传入指针，且不能为nil
// 它会把src与dst的相同字段名的值，复制到dst中
func CopyStruct(dst, src interface{}) {
	dstValue := reflect.ValueOf(dst).Elem()
	srcValue := reflect.ValueOf(src).Elem()

	for i := 0; i < srcValue.NumField(); i++ {
		srcField := srcValue.Field(i)
		srcName := srcValue.Type().Field(i).Name
		dstFieldByName := dstValue.FieldByName(srcName)

		if dstFieldByName.IsValid() {
			switch dstFieldByName.Kind() {
			case reflect.Ptr:
				switch srcField.Kind() {
				case reflect.Ptr:
					if srcField.IsNil() {
						dstFieldByName.Set(reflect.New(dstFieldByName.Type().Elem()))
					} else {
						dstFieldByName.Set(srcField)
					}
				default:
					dstFieldByName.Set(srcField.Addr())
				}
			default:
				switch srcField.Kind() {
				case reflect.Ptr:
					if srcField.IsNil() {
						dstFieldByName.Set(reflect.Zero(dstFieldByName.Type()))
					} else {
						dstFieldByName.Set(srcField.Elem())
					}
				default:
					dstFieldByName.Set(srcField)
				}
			}
		}
	}
}
