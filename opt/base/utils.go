package base

import (
	"crypto/md5"
	"fmt"
	jsoniter "github.com/json-iterator/go"
	"math/rand"
	"reflect"
	"runtime"
	"sort"
	"strings"
	"time"
)

func NewMap() map[string]interface{} {
	return make(map[string]interface{})
}

func NewMaps(length int) []map[string]interface{} {
	return make([]map[string]interface{}, length)
}

func Rand(start, end int) int {
	if start >= end {
		return end
	}
	rand.Seed(time.Now().UnixNano())
	return rand.Intn(end-start) + start
}

/*
	支持struct，map，json字符串之间互转
	@param beConverted 被转换数据
	@param to 接收转换数据
*/
func Transfer(beConverted, to interface{}) error {
	t := reflect.TypeOf(beConverted)
	if t.Name() == "string" {
		err := jsoniter.Unmarshal([]byte(beConverted.(string)), to)
		if err != nil {
			return err
		}
	} else {
		bt, err := jsoniter.Marshal(beConverted)
		if err != nil {
			return err
		}
		err = jsoniter.Unmarshal(bt, to)
		if err != nil {
			return err
		}
	}
	return nil
}

/*
	返回字符串长度（支持中文）
*/
func Length(str string) int {
	r := []rune(str)
	return len(r)
}

/*
	修改数据字段名，并返回map结构
*/
func ModifyFieldName(data interface{}, field, newField string) map[string]interface{} {
	bt, _ := jsoniter.Marshal(data)
	out := map[string]interface{}{}
	_ = jsoniter.Unmarshal(bt, &out)
	out[newField] = out[field]
	delete(out, field)
	return out
}

/*
	修改数据字段名，并返回map结构
*/
func StructToMap(data interface{}) map[string]interface{} {
	bt, _ := jsoniter.Marshal(data)
	out := map[string]interface{}{}
	_ = jsoniter.Unmarshal(bt, &out)
	return out
}

func MD5(str string) string {
	data := []byte(str)
	has := md5.Sum(data)
	md5str := fmt.Sprintf("%x", has)
	return md5str
}

func FileLine(skip int) string {
	_, path, line, _ := runtime.Caller(skip)
	i, count := len(path)-4, 0
	for ; i > 0; i-- {
		if path[i] == '/' {
			count++
			if count == 2 {
				break
			}
		}
	}
	return fmt.Sprintf("%s:%d", path[i+1:], line)
}

//验证签名
func CheckSignature(values map[string]interface{}, key string) bool {
	signName := "signature"
	sign, ok := values[signName]
	if !ok {
		return false
	}
	delete(values, signName)
	return Signature(values, key) == sign
}

//生成签名
func Signature(values map[string]interface{}, key string) string {
	pKeys := make([]string, 0)
	for key, _ := range values {
		pKeys = append(pKeys, key)
	}

	sort.Slice(pKeys, func(i, j int) bool {
		return pKeys[i] < pKeys[j]
	})
	content := make([]string, len(pKeys))
	for i, key := range pKeys {
		content[i] = fmt.Sprintf("%s=%v", key, values[key])
	}
	return MD5(fmt.Sprintf("%s%s", key, strings.Join(content, "")))
}
