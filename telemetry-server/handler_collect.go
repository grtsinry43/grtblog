package main

import (
	"encoding/json"
	"time"

	"github.com/gofiber/fiber/v2"
)

// nilGuardJSON returns a default empty JSON object if the input is nil or empty.
func nilGuardJSON(data json.RawMessage) json.RawMessage {
	if len(data) == 0 {
		return json.RawMessage("{}")
	}
	return data
}

// IncomingReport is the parsed telemetry snapshot ready for storage.
type IncomingReport struct {
	InstanceID string
	Version    string
	GoVersion  string
	OS         string
	Arch       string
	DeployMode string
	UptimeSec  int64
	Features   json.RawMessage
	Metrics    json.RawMessage
	ErrorCount int
	PanicCount int
	Payload    json.RawMessage
	Digests    []IncomingDigest
}

type IncomingDigest struct {
	Fingerprint   string
	Kind          string
	BizCode       string
	Location      string
	SampleMessage string
	Count         int64
	FirstSeen     *time.Time
	LastSeen      *time.Time
}

// incomingSnapshot mirrors the FullTelemetrySnapshot JSON structure from the blog server.
type incomingSnapshot struct {
	GeneratedAt string `json:"generatedAt"`
	Instance    struct {
		InstanceID    string          `json:"instanceId"`
		Version       string          `json:"version"`
		GoVersion     string          `json:"goVersion"`
		OS            string          `json:"os"`
		Arch          string          `json:"arch"`
		UptimeSeconds int64           `json:"uptimeSeconds"`
		DeployMode    string          `json:"deployMode"`
		Features      json.RawMessage `json:"features"`
	} `json:"instance"`
	Metrics json.RawMessage `json:"metrics"`
	Errors  []digestJSON    `json:"errors"`
	Panics  []digestJSON    `json:"panics"`
	Summary struct {
		UniqueErrors int `json:"uniqueErrors"`
		TotalErrors  int `json:"totalErrors"`
		UniquePanics int `json:"uniquePanics"`
		TotalPanics  int `json:"totalPanics"`
	} `json:"summary"`
}

type digestJSON struct {
	Fingerprint   string  `json:"fingerprint"`
	Kind          string  `json:"kind"`
	BizCode       string  `json:"bizCode"`
	Location      string  `json:"location"`
	SampleMessage string  `json:"sampleMessage"`
	Count         int64   `json:"count"`
	FirstSeen     *string `json:"firstSeen"`
	LastSeen      *string `json:"lastSeen"`
}

// CollectHandler returns a Fiber handler for POST /collect.
func CollectHandler(store *Store, limiter *RateLimiter) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Check User-Agent.
		if ua := c.Get("User-Agent"); ua != "grtblog-telemetry/1.0" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid user-agent"})
		}

		var snap incomingSnapshot
		if err := c.BodyParser(&snap); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid JSON"})
		}

		if snap.Instance.InstanceID == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "missing instanceId"})
		}

		// Rate limit.
		if !limiter.Allow(snap.Instance.InstanceID) {
			return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{"error": "rate limited, try again later"})
		}

		// Build report.
		report := &IncomingReport{
			InstanceID: snap.Instance.InstanceID,
			Version:    snap.Instance.Version,
			GoVersion:  snap.Instance.GoVersion,
			OS:         snap.Instance.OS,
			Arch:       snap.Instance.Arch,
			DeployMode: snap.Instance.DeployMode,
			UptimeSec:  snap.Instance.UptimeSeconds,
			Features:   nilGuardJSON(snap.Instance.Features),
			Metrics:    nilGuardJSON(snap.Metrics),
			ErrorCount: snap.Summary.TotalErrors,
			PanicCount: snap.Summary.TotalPanics,
			Payload:    c.Body(),
		}

		// Parse digests (errors + panics combined).
		for _, list := range [][]digestJSON{snap.Errors, snap.Panics} {
			for _, d := range list {
				dig := IncomingDigest{
					Fingerprint:   d.Fingerprint,
					Kind:          d.Kind,
					BizCode:       d.BizCode,
					Location:      d.Location,
					SampleMessage: d.SampleMessage,
					Count:         d.Count,
				}
				if d.FirstSeen != nil {
					if t, err := time.Parse(time.RFC3339, *d.FirstSeen); err == nil {
						dig.FirstSeen = &t
					}
				}
				if d.LastSeen != nil {
					if t, err := time.Parse(time.RFC3339, *d.LastSeen); err == nil {
						dig.LastSeen = &t
					}
				}
				report.Digests = append(report.Digests, dig)
			}
		}

		reportID, err := store.InsertReport(c.UserContext(), report)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "storage error"})
		}

		return c.Status(fiber.StatusAccepted).JSON(fiber.Map{
			"status":   "accepted",
			"reportId": reportID,
		})
	}
}
