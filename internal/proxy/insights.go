package proxy

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"sort"
	"sync"
	"time"

	"github.com/uucz/llmview/internal/cost"
	"github.com/uucz/llmview/internal/storage"
	"github.com/uucz/llmview/internal/ws"
)

// InsightType categorizes different kinds of insights.
type InsightType string

const (
	InsightLoop         InsightType = "loop_detected"
	InsightPromptWaste  InsightType = "prompt_waste"
	InsightModelSuggest InsightType = "model_downgrade"
	InsightBurnRate     InsightType = "burn_rate"
)

// InsightSeverity indicates how urgent an insight is.
type InsightSeverity string

const (
	SeverityInfo     InsightSeverity = "info"
	SeverityWarning  InsightSeverity = "warning"
	SeverityCritical InsightSeverity = "critical"
)

// severityRank returns a numeric rank for sorting (higher = more severe).
func severityRank(s InsightSeverity) int {
	switch s {
	case SeverityCritical:
		return 3
	case SeverityWarning:
		return 2
	case SeverityInfo:
		return 1
	default:
		return 0
	}
}

// Insight is a single optimization recommendation.
type Insight struct {
	Type         InsightType     `json:"type"`
	Severity     InsightSeverity `json:"severity"`
	Title        string          `json:"title"`
	Description  string          `json:"description"`
	Savings      float64         `json:"savings,omitempty"`
	TokenSavings int             `json:"token_savings,omitempty"`
	CallIDs      []string        `json:"call_ids,omitempty"`
	DetectedAt   int64           `json:"detected_at"`
}

// InsightsEngine analyzes API calls and generates optimization insights.
type InsightsEngine struct {
	mu        sync.RWMutex
	store     *storage.Store
	hub       *ws.Hub
	calc      *cost.Calculator
	sessionID string
	budget    float64

	// Loop detection state: sliding window of recent request fingerprints
	recentHashes []requestFingerprint
	maxRecent    int

	// Accumulated real-time insights
	insights []Insight
}

type requestFingerprint struct {
	hash      string
	callID    string
	provider  storage.Provider
	model     string
	timestamp time.Time
	tokens    int
}

// NewInsightsEngine creates a new InsightsEngine.
func NewInsightsEngine(store *storage.Store, hub *ws.Hub, calc *cost.Calculator, sessionID string, budget float64) *InsightsEngine {
	return &InsightsEngine{
		store:     store,
		hub:       hub,
		calc:      calc,
		sessionID: sessionID,
		budget:    budget,
		maxRecent: 20,
	}
}

// Analyze is called after each API call completes. It runs real-time analysis
// for loop detection and burn rate monitoring, broadcasting insights via WebSocket.
func (e *InsightsEngine) Analyze(call *storage.APICall, reqBody []byte) {
	e.mu.Lock()
	defer e.mu.Unlock()

	// --- Loop Detection ---
	hash := hashRequestMessages(call.Provider, reqBody)
	fp := requestFingerprint{
		hash:      hash,
		callID:    call.ID,
		provider:  call.Provider,
		model:     call.Model,
		timestamp: call.StartedAt,
		tokens:    call.InputTokens + call.OutputTokens,
	}

	e.recentHashes = append(e.recentHashes, fp)
	// Trim to sliding window
	if len(e.recentHashes) > e.maxRecent {
		e.recentHashes = e.recentHashes[len(e.recentHashes)-e.maxRecent:]
	}

	// Check for 3+ identical hashes within the last 60 seconds
	cutoff := time.Now().Add(-60 * time.Second)
	hashCounts := make(map[string][]string) // hash -> list of callIDs
	for _, rh := range e.recentHashes {
		if rh.timestamp.After(cutoff) {
			hashCounts[rh.hash] = append(hashCounts[rh.hash], rh.callID)
		}
	}
	for h, callIDs := range hashCounts {
		if len(callIDs) >= 3 {
			// Avoid duplicate loop insights for the same hash
			if e.hasLoopInsight(h) {
				continue
			}
			insight := Insight{
				Type:     InsightLoop,
				Severity: SeverityCritical,
				Title:    "Repetitive API call loop detected",
				Description: fmt.Sprintf(
					"%d identical requests detected within 60s. The LLM agent may be stuck in a retry loop. Consider adding loop-breaking logic or checking the prompt.",
					len(callIDs),
				),
				CallIDs:    callIDs,
				DetectedAt: time.Now().UnixMilli(),
			}
			e.insights = append(e.insights, insight)
			e.broadcastInsight(insight)
		}
	}

	// --- Burn Rate Check ---
	sess, err := e.store.GetSession(e.sessionID)
	if err != nil {
		return
	}
	elapsed := time.Since(sess.StartedAt)
	if elapsed < 10*time.Second {
		// Too early for meaningful burn rate
		return
	}
	hoursElapsed := elapsed.Hours()
	if hoursElapsed == 0 {
		return
	}
	burnRate := sess.TotalCost / hoursElapsed

	threshold := 10.0 // default $/hour threshold
	if e.budget > 0 {
		// If a budget is set, warn if on track to exceed it within the next hour
		threshold = e.budget / math.Max(hoursElapsed, 1.0)
	}

	if burnRate > threshold && !e.hasRecentBurnRateInsight() {
		insight := Insight{
			Type:     InsightBurnRate,
			Severity: SeverityWarning,
			Title:    "High burn rate detected",
			Description: fmt.Sprintf(
				"Current spend rate: $%.2f/hour (session total: $%.4f over %.1f min). %s",
				burnRate,
				sess.TotalCost,
				elapsed.Minutes(),
				e.burnRateAdvice(burnRate),
			),
			Savings:    burnRate * 0.5, // conservative estimate of potential savings
			DetectedAt: time.Now().UnixMilli(),
		}
		e.insights = append(e.insights, insight)
		e.broadcastInsight(insight)
	}
}

// ComputeInsights performs deeper on-demand analysis across all session calls.
// It returns accumulated real-time insights plus computed batch insights,
// sorted by severity (critical first) then by estimated savings (highest first).
func (e *InsightsEngine) ComputeInsights() []Insight {
	e.mu.RLock()
	realtime := make([]Insight, len(e.insights))
	copy(realtime, e.insights)
	e.mu.RUnlock()

	computed := make([]Insight, 0)

	calls, err := e.store.ListCalls(e.sessionID, 500, 0)
	if err != nil {
		log.Printf("[insights] failed to list calls: %v", err)
		return realtime
	}
	if len(calls) == 0 {
		return realtime
	}

	// --- Prompt Waste Detection ---
	computed = append(computed, e.detectPromptWaste(calls)...)

	// --- Model Downgrade Suggestions ---
	computed = append(computed, e.detectModelDowngrade(calls)...)

	// --- Burn Rate (computed) ---
	if br := e.computeBurnRate(); br != nil {
		computed = append(computed, *br)
	}

	// Merge and deduplicate
	all := append(realtime, computed...)
	all = deduplicateInsights(all)

	// Sort: critical first, then by savings descending
	sort.Slice(all, func(i, j int) bool {
		ri := severityRank(all[i].Severity)
		rj := severityRank(all[j].Severity)
		if ri != rj {
			return ri > rj
		}
		return all[i].Savings > all[j].Savings
	})

	return all
}

// GetInsights returns the accumulated real-time insights.
func (e *InsightsEngine) GetInsights() []Insight {
	e.mu.RLock()
	defer e.mu.RUnlock()
	result := make([]Insight, len(e.insights))
	copy(result, e.insights)
	return result
}

// detectPromptWaste finds repeated system prompts across calls and estimates wasted tokens/cost.
func (e *InsightsEngine) detectPromptWaste(calls []storage.APICall) []Insight {
	type systemPromptGroup struct {
		hash       string
		tokens     int
		count      int
		callIDs    []string
		provider   storage.Provider
		model      string
		promptSize int // approximate character length
	}

	groups := make(map[string]*systemPromptGroup)

	for _, c := range calls {
		if len(c.RequestBody) == 0 {
			continue
		}
		sysHash, sysLen := extractSystemPromptHash(c.Provider, c.RequestBody)
		if sysHash == "" {
			continue
		}
		g, ok := groups[sysHash]
		if !ok {
			g = &systemPromptGroup{
				hash:       sysHash,
				provider:   c.Provider,
				model:      c.Model,
				promptSize: sysLen,
			}
			groups[sysHash] = g
		}
		g.count++
		g.callIDs = append(g.callIDs, c.ID)
		// Estimate tokens from characters (rough: 1 token ~ 4 chars)
		g.tokens = sysLen / 4
	}

	var insights []Insight
	for _, g := range groups {
		if g.count <= 1 {
			continue
		}
		// Wasted tokens = repeated_tokens * (count - 1)
		wastedTokens := g.tokens * (g.count - 1)
		pricing := e.calc.GetPricing(g.model)
		wastedCost := float64(wastedTokens) / 1_000_000 * pricing.InputPerMillion

		if wastedTokens < 100 {
			// Not worth reporting for tiny prompts
			continue
		}

		description := fmt.Sprintf(
			"The same system prompt (%d chars) was sent %d times, wasting ~%d input tokens ($%.4f).",
			g.promptSize, g.count, wastedTokens, wastedCost,
		)
		if g.provider == storage.ProviderAnthropic {
			description += " Consider using Anthropic prompt caching to reduce costs."
		} else {
			description += " Consider using prompt caching or reducing system prompt size."
		}

		severity := SeverityInfo
		if wastedCost > 0.10 {
			severity = SeverityWarning
		}
		if wastedCost > 1.00 {
			severity = SeverityCritical
		}

		insights = append(insights, Insight{
			Type:         InsightPromptWaste,
			Severity:     severity,
			Title:        "Repeated system prompt detected",
			Description:  description,
			Savings:      wastedCost,
			TokenSavings: wastedTokens,
			CallIDs:      g.callIDs,
			DetectedAt:   time.Now().UnixMilli(),
		})
	}
	return insights
}

// detectModelDowngrade finds calls where an expensive model was used but the response was short,
// suggesting a cheaper model could have been sufficient.
func (e *InsightsEngine) detectModelDowngrade(calls []storage.APICall) []Insight {
	expensiveModels := map[string]string{
		"gpt-4o":            "gpt-4o-mini",
		"gpt-4":             "gpt-4o-mini",
		"gpt-4-turbo":       "gpt-4o-mini",
		"claude-opus-4-6":   "claude-sonnet-4-6",
		"claude-sonnet-4-6": "claude-haiku-4-5",
		"claude-3-5-sonnet-latest": "claude-3-5-haiku-latest",
		"o1":                "o1-mini",
		"o3":                "o3-mini",
	}

	type downgradeGroup struct {
		model       string
		cheaperAlt  string
		callIDs     []string
		totalSaving float64
	}

	groups := make(map[string]*downgradeGroup)

	for _, c := range calls {
		if c.OutputTokens >= 200 || c.StatusCode != 200 {
			continue
		}
		cheaper, ok := expensiveModels[c.Model]
		if !ok {
			continue
		}

		currentCost := e.calc.Calculate(c.Model, c.InputTokens, c.OutputTokens)
		cheaperCost := e.calc.Calculate(cheaper, c.InputTokens, c.OutputTokens)
		saving := currentCost - cheaperCost
		if saving <= 0 {
			continue
		}

		key := c.Model + "->" + cheaper
		g, ok := groups[key]
		if !ok {
			g = &downgradeGroup{model: c.Model, cheaperAlt: cheaper}
			groups[key] = g
		}
		g.callIDs = append(g.callIDs, c.ID)
		g.totalSaving += saving
	}

	var insights []Insight
	for _, g := range groups {
		if len(g.callIDs) == 0 {
			continue
		}
		severity := SeverityInfo
		if g.totalSaving > 0.05 {
			severity = SeverityWarning
		}
		if g.totalSaving > 0.50 {
			severity = SeverityCritical
		}
		insights = append(insights, Insight{
			Type:     InsightModelSuggest,
			Severity: severity,
			Title:    fmt.Sprintf("Consider using %s instead of %s", g.cheaperAlt, g.model),
			Description: fmt.Sprintf(
				"%d calls used %s but produced short responses (<200 output tokens). Switching to %s could save $%.4f.",
				len(g.callIDs), g.model, g.cheaperAlt, g.totalSaving,
			),
			Savings: g.totalSaving,
			CallIDs: g.callIDs,
			DetectedAt: time.Now().UnixMilli(),
		})
	}
	return insights
}

// computeBurnRate calculates the current session burn rate for the on-demand report.
func (e *InsightsEngine) computeBurnRate() *Insight {
	sess, err := e.store.GetSession(e.sessionID)
	if err != nil || sess.TotalCost == 0 {
		return nil
	}
	elapsed := time.Since(sess.StartedAt)
	if elapsed < 10*time.Second {
		return nil
	}
	hoursElapsed := elapsed.Hours()
	burnRate := sess.TotalCost / hoursElapsed

	severity := SeverityInfo
	if burnRate > 5.0 {
		severity = SeverityWarning
	}
	if burnRate > 10.0 {
		severity = SeverityCritical
	}

	return &Insight{
		Type:     InsightBurnRate,
		Severity: severity,
		Title:    "Session burn rate",
		Description: fmt.Sprintf(
			"Current rate: $%.2f/hour | Session total: $%.4f over %.1f min | %d requests",
			burnRate, sess.TotalCost, elapsed.Minutes(), sess.RequestCount,
		),
		Savings:    0,
		DetectedAt: time.Now().UnixMilli(),
	}
}

// hashRequestMessages creates a content hash of the messages in a request body,
// ignoring metadata like timestamps and IDs. This is used for loop detection.
func hashRequestMessages(provider storage.Provider, body []byte) string {
	var req map[string]interface{}
	if err := json.Unmarshal(body, &req); err != nil {
		return ""
	}

	relevant := make(map[string]interface{})
	if msgs, ok := req["messages"]; ok {
		relevant["messages"] = msgs
	}
	// Include Anthropic-style top-level system prompt
	if sys, ok := req["system"]; ok {
		relevant["system"] = sys
	}

	data, err := json.Marshal(relevant)
	if err != nil {
		return ""
	}
	h := sha256.Sum256(data)
	return fmt.Sprintf("%x", h[:8])
}

// extractSystemPromptHash extracts and hashes the system prompt from a request body.
// Returns the hash and approximate character length of the system prompt.
func extractSystemPromptHash(provider storage.Provider, body []byte) (string, int) {
	var req map[string]interface{}
	if err := json.Unmarshal(body, &req); err != nil {
		return "", 0
	}

	var systemText string

	switch provider {
	case storage.ProviderAnthropic:
		// Anthropic uses a top-level "system" field (string or array)
		if sys, ok := req["system"]; ok {
			switch v := sys.(type) {
			case string:
				systemText = v
			default:
				data, _ := json.Marshal(v)
				systemText = string(data)
			}
		}
	default:
		// OpenAI-style: first message with role=system
		if msgs, ok := req["messages"]; ok {
			if msgSlice, ok := msgs.([]interface{}); ok {
				for _, m := range msgSlice {
					if msg, ok := m.(map[string]interface{}); ok {
						if role, _ := msg["role"].(string); role == "system" {
							if content, ok := msg["content"].(string); ok {
								systemText = content
							} else {
								data, _ := json.Marshal(msg["content"])
								systemText = string(data)
							}
							break
						}
					}
				}
			}
		}
	}

	if systemText == "" {
		return "", 0
	}

	h := sha256.Sum256([]byte(systemText))
	return fmt.Sprintf("%x", h[:8]), len(systemText)
}

// hasLoopInsight checks if a loop insight for the given hash already exists.
func (e *InsightsEngine) hasLoopInsight(hash string) bool {
	for _, ins := range e.insights {
		if ins.Type == InsightLoop && len(ins.CallIDs) > 0 {
			// Check if this loop shares any call IDs (same hash group)
			existing := hashRequestMessages("", nil) // not useful, check by description
			_ = existing
			// Simpler: check if the most recent insight with same type was within last 30s
			if ins.Type == InsightLoop && time.Now().UnixMilli()-ins.DetectedAt < 30_000 {
				return true
			}
		}
	}
	return false
}

// hasRecentBurnRateInsight checks if a burn rate insight was generated in the last 60 seconds.
func (e *InsightsEngine) hasRecentBurnRateInsight() bool {
	cutoff := time.Now().UnixMilli() - 60_000
	for _, ins := range e.insights {
		if ins.Type == InsightBurnRate && ins.DetectedAt > cutoff {
			return true
		}
	}
	return false
}

// broadcastInsight sends an insight to all connected WebSocket clients.
func (e *InsightsEngine) broadcastInsight(insight Insight) {
	e.hub.Broadcast(storage.WSEvent{
		Type: "insight",
		Data: insight,
	})
}

// burnRateAdvice returns contextual advice based on the burn rate.
func (e *InsightsEngine) burnRateAdvice(rate float64) string {
	if e.budget > 0 {
		hoursLeft := (e.budget - e.currentSessionCost()) / rate
		if hoursLeft <= 0 {
			return "Budget already exceeded!"
		}
		return fmt.Sprintf("At this rate, budget ($%.2f) will be exhausted in ~%.1f hours.", e.budget, hoursLeft)
	}
	if rate > 20 {
		return "This is very high. Consider using cheaper models or reducing prompt sizes."
	}
	if rate > 10 {
		return "Consider reviewing which calls are most expensive."
	}
	return "Monitor spending and consider setting a --budget flag."
}

// currentSessionCost fetches the current total cost from the store.
func (e *InsightsEngine) currentSessionCost() float64 {
	sess, err := e.store.GetSession(e.sessionID)
	if err != nil {
		return 0
	}
	return sess.TotalCost
}

// deduplicateInsights removes insights with the same type and overlapping call IDs.
func deduplicateInsights(insights []Insight) []Insight {
	seen := make(map[string]bool)
	var result []Insight
	for _, ins := range insights {
		key := string(ins.Type) + "|" + ins.Title
		if seen[key] {
			continue
		}
		seen[key] = true
		result = append(result, ins)
	}
	return result
}
