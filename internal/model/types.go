package model

import (
	"sort"
	"strings"
	"time"
)

type PolicyStatus string

const (
	PolicyDraft   PolicyStatus = "draft"
	PolicyActive  PolicyStatus = "active"
	PolicyRetired PolicyStatus = "retired"
)

type ComponentStatus string

const (
	ComponentPending  ComponentStatus = "pending"
	ComponentTrusted  ComponentStatus = "trusted"
	ComponentDisputed ComponentStatus = "disputed"
	ComponentRevoked  ComponentStatus = "revoked"
)

type AnalysisStatus string

const (
	AnalysisQueued     AnalysisStatus = "queued"
	AnalysisRunning    AnalysisStatus = "running"
	AnalysisBlocked    AnalysisStatus = "blocked"
	AnalysisReviewable AnalysisStatus = "reviewable"
	AnalysisPublished  AnalysisStatus = "published"
	AnalysisSuperseded AnalysisStatus = "superseded"
)

type FindingKind string

const (
	FindingObligation FindingKind = "obligation"
	FindingBlocker    FindingKind = "blocker"
	FindingWarning    FindingKind = "warning"
	FindingWaived     FindingKind = "waived"
)

type WaiverStatus string

const (
	WaiverRequested WaiverStatus = "requested"
	WaiverApproved  WaiverStatus = "approved"
	WaiverRejected  WaiverStatus = "rejected"
	WaiverExpired   WaiverStatus = "expired"
)

type PolicyRule struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Forbidden   []string `json:"forbidden"`
	RequireNote []string `json:"require_note"`
	AllowWaiver bool     `json:"allow_waiver"`
}

type Policy struct {
	ID        string       `json:"id"`
	Name      string       `json:"name"`
	Version   int          `json:"version"`
	Status    PolicyStatus `json:"status"`
	Rules     []PolicyRule `json:"rules"`
	CreatedAt time.Time    `json:"created_at"`
	UpdatedAt time.Time    `json:"updated_at"`
}

type Component struct {
	ID       string            `json:"id"`
	Name     string            `json:"name"`
	Version  string            `json:"version"`
	License  string            `json:"license"`
	Source   string            `json:"source"`
	Status   ComponentStatus   `json:"status"`
	Metadata map[string]string `json:"metadata,omitempty"`
}

type Edge struct {
	From  string `json:"from"`
	To    string `json:"to"`
	Scope string `json:"scope"`
}

type Submission struct {
	ID          string      `json:"id"`
	Name        string      `json:"name"`
	Release     string      `json:"release"`
	Fingerprint string      `json:"fingerprint"`
	Components  []Component `json:"components"`
	Edges       []Edge      `json:"edges"`
	CreatedAt   time.Time   `json:"created_at"`
}

type Finding struct {
	ID         string      `json:"id"`
	AnalysisID string      `json:"analysis_id"`
	Kind       FindingKind `json:"kind"`
	Code       string      `json:"code"`
	Message    string      `json:"message"`
	Components []string    `json:"components"`
	Derivation []string    `json:"derivation"`
	WaiverID   string      `json:"waiver_id,omitempty"`
	CreatedAt  time.Time   `json:"created_at"`
}

type Analysis struct {
	ID             string         `json:"id"`
	SubmissionID   string         `json:"submission_id"`
	PolicyID       string         `json:"policy_id"`
	Status         AnalysisStatus `json:"status"`
	InputSnapshot  string         `json:"input_snapshot"`
	PolicySnapshot string         `json:"policy_snapshot"`
	PublishedAt    *time.Time     `json:"published_at,omitempty"`
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`
}

type Waiver struct {
	ID         string       `json:"id"`
	AnalysisID string       `json:"analysis_id"`
	FindingID  string       `json:"finding_id"`
	Reason     string       `json:"reason"`
	Status     WaiverStatus `json:"status"`
	ExpiresAt  *time.Time   `json:"expires_at,omitempty"`
	CreatedAt  time.Time    `json:"created_at"`
	UpdatedAt  time.Time    `json:"updated_at"`
}

type Decision struct {
	Analysis Analysis  `json:"analysis"`
	Findings []Finding `json:"findings"`
}

func FindingKey(f Finding) string {
	components := append([]string(nil), f.Components...)
	sort.Strings(components)
	return f.Code + "|" + strings.Join(components, ",") + "|" + f.Message
}

func ValidateComparisonIDs(left, right string) error {
	return nil
}
