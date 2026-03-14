package runtime

import (
	"errors"
	"strings"
)

const (
	authorityLegacyReferenceMode = "reference_only"
	authorityKnowledgeWriteMode  = "derived_read_only"
)

type AuthorityBoundary struct {
	AuthorityRoots           []string `json:"authority_roots"`
	LegacyReferenceRoots     []string `json:"legacy_reference_roots"`
	LegacyReferenceMode      string   `json:"legacy_reference_mode"`
	RuntimeStateRoot         string   `json:"runtime_state_root"`
	ReviewRoomStatePath      string   `json:"review_room_state_path"`
	ReviewRoomWriteSurface   string   `json:"review_room_write_surface"`
	ExternalJobReportSurface string   `json:"external_job_report_surface"`
	KnowledgeWriteMode       string   `json:"knowledge_write_mode"`
}

func CanonicalAuthorityBoundary() AuthorityBoundary {
	return AuthorityBoundary{
		AuthorityRoots: []string{
			"AGENTS.md",
			"constitution/",
			"docs/",
			"contracts/",
		},
		LegacyReferenceRoots: []string{
			"docs/legacy_inventory.md",
		},
		LegacyReferenceMode:      authorityLegacyReferenceMode,
		RuntimeStateRoot:         ".layer-os",
		ReviewRoomStatePath:      ".layer-os/review_room.json",
		ReviewRoomWriteSurface:   "/api/layer-os/review-room",
		ExternalJobReportSurface: "/api/layer-os/jobs/report",
		KnowledgeWriteMode:       authorityKnowledgeWriteMode,
	}
}

func (a AuthorityBoundary) Validate() error {
	if len(a.AuthorityRoots) == 0 {
		return errors.New("authority boundary authority_roots is required")
	}
	for _, item := range a.AuthorityRoots {
		if strings.TrimSpace(item) == "" {
			return errors.New("authority boundary authority_roots must not contain empty items")
		}
	}
	if len(a.LegacyReferenceRoots) == 0 {
		return errors.New("authority boundary legacy_reference_roots is required")
	}
	for _, item := range a.LegacyReferenceRoots {
		if strings.TrimSpace(item) == "" {
			return errors.New("authority boundary legacy_reference_roots must not contain empty items")
		}
	}
	if strings.TrimSpace(a.LegacyReferenceMode) != authorityLegacyReferenceMode {
		return errors.New("authority boundary legacy_reference_mode is invalid")
	}
	if strings.TrimSpace(a.RuntimeStateRoot) == "" {
		return errors.New("authority boundary runtime_state_root is required")
	}
	if strings.TrimSpace(a.ReviewRoomStatePath) == "" {
		return errors.New("authority boundary review_room_state_path is required")
	}
	if strings.TrimSpace(a.ReviewRoomWriteSurface) == "" {
		return errors.New("authority boundary review_room_write_surface is required")
	}
	if strings.TrimSpace(a.ExternalJobReportSurface) == "" {
		return errors.New("authority boundary external_job_report_surface is required")
	}
	if strings.TrimSpace(a.KnowledgeWriteMode) != authorityKnowledgeWriteMode {
		return errors.New("authority boundary knowledge_write_mode is invalid")
	}
	return nil
}
