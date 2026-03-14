package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"layer-os/internal/runtime"
)

func TestProposalRoutesCreateAndPromote(t *testing.T) {
	service, err := runtime.NewService(t.TempDir())
	if err != nil {
		t.Fatalf("new service: %v", err)
	}
	router := NewRouter(service)

	createRaw := []byte(`{"proposal_id":"proposal_001","title":"Plan queue","intent":"close planning gap","summary":"Plan queue","surface":"api","priority":"high","risk":"medium","status":"proposed","notes":[]}`)
	createReq := httptest.NewRequest(http.MethodPost, "/api/layer-os/proposals", bytes.NewReader(createRaw))
	createReq.Header.Set("Content-Type", "application/json")
	createRec := httptest.NewRecorder()
	router.ServeHTTP(createRec, createReq)
	if createRec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", createRec.Code, createRec.Body.String())
	}

	promoteRaw := []byte(`{"proposal_id":"proposal_001","work_item_id":"work_001"}`)
	promoteReq := httptest.NewRequest(http.MethodPost, "/api/layer-os/proposals/promote", bytes.NewReader(promoteRaw))
	promoteReq.Header.Set("Content-Type", "application/json")
	promoteRec := httptest.NewRecorder()
	router.ServeHTTP(promoteRec, promoteReq)
	if promoteRec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", promoteRec.Code, promoteRec.Body.String())
	}
	var response struct {
		Proposal runtime.ProposalItem `json:"proposal"`
		WorkItem runtime.WorkItem     `json:"work_item"`
	}
	if err := json.NewDecoder(promoteRec.Body).Decode(&response); err != nil {
		t.Fatalf("decode promote response: %v", err)
	}
	if response.Proposal.Status != "promoted" || response.WorkItem.ID != "work_001" {
		t.Fatalf("unexpected promote response: %+v", response)
	}
}
