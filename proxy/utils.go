package main

import (
	"fmt"
	"kadam.net/test_task"
	"strings"
)

var (
	headerContentTypeJson = []byte("application/json")
)

func parseRecipients(rawRecipients string) []string {
	ports := strings.Split(rawRecipients, ",")

	recipients := make([]string, len(ports))

	for i, p := range ports {
		recipients[i] = fmt.Sprintf("http://localhost:%s%s", p, test_task.BidEndpoint)
	}

	return recipients
}

// commonToInnerRequest converts CommonProxyRequest to InnerRequest with a little hint:
// The only fields that matter are Id and MinPrice, so we only set them.
func commonToInnerRequest(common test_task.CommonProxyRequest) test_task.InnerRequest {
	var id string
	switch true {
	case common.Id != nil:
		id = *common.Id
	case common.Key != nil:
		id = fmt.Sprintf("%d", *common.Key)
	case common.Name != nil:
		id = *common.Name
	}

	var minPrice float64
	switch true {
	case common.Balance != nil:
		minPrice = *common.Balance
	case common.Price != nil:
		minPrice = float64(*common.Price)
	case common.Bid != nil:
		minPrice = float64(*common.Bid)
	}

	return test_task.InnerRequest{
		Id:       id,
		MinPrice: minPrice,
	}
}
