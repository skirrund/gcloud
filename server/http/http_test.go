package http

import (
	"fmt"
	"testing"

	"github.com/skirrund/gcloud/server/lb"
)

func TestGet(t *testing.T) {
	var resp []byte
	lb.GetInstance().SetHttpClient(lb.FastHttpClient{})
	params := map[string]string{"code": "5a8fdab9a0ddfc3c0d5d16331dd394ab1d21d3e66c365c694fdea57bd8b38fca"}
	PostJSONUrl("https://api.weixin.qq.com/wxa/business/getuserphonenumber?access_token=71_mpMPZl0rKXP-HWjF5iDHBnqHVgRvgoA0wH85GxXQewm8-J45gv4PMG49gW6llKeIfWjQNgIXbeA5nMYpkVvE8p8MTHQuYxmocK7wbq2Z3v1gLFLh3ZAvSHLFOnkLMIaAJAABX", nil, params, &resp)

	//PostJSONUrl("http://127.0.0.1:8080/test?ac=testsetsetset", nil, params, &resp)
	fmt.Println(string(resp))
}
