package call

import (
	"strconv"
	"strings"
	"sync/atomic"
	"time"

	jsoniter "github.com/json-iterator/go"
)

var number uint64 = 0

func getMessageId() string {
	var str strings.Builder
	str.WriteString(serviceName)
	str.WriteByte(':')
	str.WriteString(serviceNo)
	str.WriteByte('_')
	str.WriteByte('m')
	str.WriteByte('_')
	str.WriteString(strconv.FormatInt(time.Now().Unix(), 10))
	str.WriteByte('_')
	str.WriteString(strconv.FormatUint(atomic.AddUint64(&number, 1), 10))
	return str.String()
}

func getRequestId() string {
	var str strings.Builder
	str.WriteString(serviceName)
	str.WriteByte(':')
	str.WriteString(serviceNo)
	str.WriteByte('_')
	str.WriteByte('r')
	str.WriteByte('_')
	str.WriteString(strconv.FormatInt(time.Now().Unix(), 10))
	str.WriteByte('_')
	str.WriteString(strconv.FormatUint(atomic.AddUint64(&number, 1), 10))
	return str.String()
}

func encode(data any) (r []byte) {
	var json = jsoniter.ConfigCompatibleWithStandardLibrary
	r, _ = json.Marshal(&data)
	return
}

func decode(data []byte, v any) error {
	var json = jsoniter.ConfigCompatibleWithStandardLibrary
	return json.Unmarshal(data, &v)
}
