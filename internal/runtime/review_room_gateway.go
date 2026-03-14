package runtime

func reviewRoomSuggestionForGatewayCall(item GatewayCall) (ReviewRoomItem, bool) {
	if item.Status != "failed" {
		return ReviewRoomItem{}, false
	}
	ref := item.CallID
	evidence := []string{"gateway:" + item.CallID, "provider:" + item.Provider}
	if item.LastError != nil && *item.LastError != "" {
		evidence = append(evidence, "error:"+*item.LastError)
	}
	if item.LastHTTPStatus != nil {
		evidence = append(evidence, "http_status:"+strconvItoa(*item.LastHTTPStatus))
	}
	return newSignalReviewRoomItem("게이트웨이 호출 `"+item.CallID+"`이 공급자 `"+item.Provider+"`에서 실패했어. 이 레인을 자동화가 믿고 이어가기 전에 founder 검토가 필요해.", "gateway.failed", &ref, "failed provider dispatch requires review before agent automation depends on the provider lane", "review_room.auto.gateway_failed", evidence), true
}
