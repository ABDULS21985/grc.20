package service

import (
	"testing"

	"github.com/google/uuid"
)

// ============================================================
// detectCycles Tests — DFS-based circular dependency detection
// ============================================================

func TestDetectCycles_NoCycles(t *testing.T) {
	// A -> B -> C (linear, no cycle)
	a := uuid.New()
	b := uuid.New()
	c := uuid.New()

	adjList := map[uuid.UUID][]uuid.UUID{
		a: {b},
		b: {c},
	}

	cycles := detectCycles(adjList)
	if len(cycles) != 0 {
		t.Errorf("Expected no cycles in linear graph, got %d cycles", len(cycles))
	}
}

func TestDetectCycles_SimpleCycle(t *testing.T) {
	// A -> B -> C -> A (triangle cycle)
	a := uuid.New()
	b := uuid.New()
	c := uuid.New()

	adjList := map[uuid.UUID][]uuid.UUID{
		a: {b},
		b: {c},
		c: {a},
	}

	cycles := detectCycles(adjList)
	if len(cycles) == 0 {
		t.Error("Expected to detect a cycle in A -> B -> C -> A, got none")
	}

	// Verify the cycle contains the right nodes
	found := false
	for _, cycle := range cycles {
		if len(cycle) >= 3 {
			nodeSet := make(map[uuid.UUID]bool)
			for _, n := range cycle {
				nodeSet[n] = true
			}
			if nodeSet[a] && nodeSet[b] && nodeSet[c] {
				found = true
				break
			}
		}
	}
	if !found {
		t.Error("Detected cycle does not contain expected nodes A, B, C")
	}
}

func TestDetectCycles_SelfLoop(t *testing.T) {
	// A -> A (self-referencing)
	a := uuid.New()

	adjList := map[uuid.UUID][]uuid.UUID{
		a: {a},
	}

	cycles := detectCycles(adjList)
	if len(cycles) == 0 {
		t.Error("Expected to detect a self-loop cycle, got none")
	}
}

func TestDetectCycles_MultipleCycles(t *testing.T) {
	// Two independent cycles:
	//   A -> B -> A
	//   C -> D -> E -> C
	a := uuid.New()
	b := uuid.New()
	c := uuid.New()
	d := uuid.New()
	e := uuid.New()

	adjList := map[uuid.UUID][]uuid.UUID{
		a: {b},
		b: {a},
		c: {d},
		d: {e},
		e: {c},
	}

	cycles := detectCycles(adjList)
	if len(cycles) < 2 {
		t.Errorf("Expected at least 2 cycles, got %d", len(cycles))
	}
}

func TestDetectCycles_DAG(t *testing.T) {
	// Diamond DAG: A -> B, A -> C, B -> D, C -> D (no cycle)
	a := uuid.New()
	b := uuid.New()
	c := uuid.New()
	d := uuid.New()

	adjList := map[uuid.UUID][]uuid.UUID{
		a: {b, c},
		b: {d},
		c: {d},
	}

	cycles := detectCycles(adjList)
	if len(cycles) != 0 {
		t.Errorf("Expected no cycles in DAG, got %d cycles", len(cycles))
	}
}

func TestDetectCycles_EmptyGraph(t *testing.T) {
	adjList := map[uuid.UUID][]uuid.UUID{}

	cycles := detectCycles(adjList)
	if len(cycles) != 0 {
		t.Errorf("Expected no cycles in empty graph, got %d", len(cycles))
	}
}

func TestDetectCycles_DisconnectedWithOneCycle(t *testing.T) {
	// Disconnected: A -> B (no cycle), C -> D -> C (cycle)
	a := uuid.New()
	b := uuid.New()
	c := uuid.New()
	d := uuid.New()

	adjList := map[uuid.UUID][]uuid.UUID{
		a: {b},
		c: {d},
		d: {c},
	}

	cycles := detectCycles(adjList)
	if len(cycles) == 0 {
		t.Error("Expected to detect cycle in disconnected component, got none")
	}

	// Verify the cycle involves C and D, not A and B
	for _, cycle := range cycles {
		for _, n := range cycle {
			if n == a || n == b {
				t.Error("Cycle incorrectly includes non-cyclic nodes A or B")
			}
		}
	}
}

// ============================================================
// SPoF Detection Logic Tests — in-memory simulation
// ============================================================

// TestSPOFAggregation verifies that dependency aggregation across
// multiple critical processes correctly identifies shared dependencies.
func TestSPOFAggregation(t *testing.T) {
	// Simulate the SPoF detection logic:
	// Process A (critical) depends on: Database X, Network Y
	// Process B (critical) depends on: Database X, Storage Z
	// Process C (non-critical) depends on: Database X
	//
	// Expected SPoF: Database X (2 critical processes depend on it)
	// Network Y and Storage Z are NOT SPoFs (only 1 critical process each)

	type dep struct {
		processID      string
		depName        string
		isCritProcess  bool
	}

	deps := []dep{
		{"procA", "DatabaseX", true},
		{"procA", "NetworkY", true},
		{"procB", "DatabaseX", true},
		{"procB", "StorageZ", true},
		{"procC", "DatabaseX", false},
	}

	// Aggregate: depName -> set of critical process IDs
	depToCritProcesses := make(map[string]map[string]bool)
	for _, d := range deps {
		if !d.isCritProcess {
			continue
		}
		if _, ok := depToCritProcesses[d.depName]; !ok {
			depToCritProcesses[d.depName] = make(map[string]bool)
		}
		depToCritProcesses[d.depName][d.processID] = true
	}

	// Filter: entities with 2+ critical processes
	var spofNames []string
	for name, procs := range depToCritProcesses {
		if len(procs) >= 2 {
			spofNames = append(spofNames, name)
		}
	}

	if len(spofNames) != 1 {
		t.Fatalf("Expected exactly 1 SPoF, got %d: %v", len(spofNames), spofNames)
	}
	if spofNames[0] != "DatabaseX" {
		t.Errorf("Expected SPoF to be DatabaseX, got %s", spofNames[0])
	}
}

// TestSPOFTransitiveDependencies verifies transitive SPoF detection:
// Process A depends on Process B, Process B depends on Asset X.
// Both A and B are critical.
// Asset X should be identified as a SPoF for both A and B.
func TestSPOFTransitiveDependencies(t *testing.T) {
	type dep struct {
		processID string
		depType   string // "process" or "asset"
		depName   string
		depTarget string // for process deps, the target process ID
	}

	// Process A -> Process B (process dependency)
	// Process B -> Asset X (asset dependency)
	allDeps := map[string][]dep{
		"procA": {
			{processID: "procA", depType: "process", depName: "ProcessB", depTarget: "procB"},
		},
		"procB": {
			{processID: "procB", depType: "system", depName: "AssetX", depTarget: ""},
		},
	}

	critProcesses := map[string]bool{"procA": true, "procB": true}

	// Simulate transitive resolution
	// For each critical process, BFS through process dependencies
	entityToProcesses := make(map[string]map[string]bool)

	for procID := range critProcesses {
		visited := make(map[string]bool)
		queue := []string{procID}

		for len(queue) > 0 {
			current := queue[0]
			queue = queue[1:]

			if visited[current] {
				continue
			}
			visited[current] = true

			for _, d := range allDeps[current] {
				if d.depType == "process" && d.depTarget != "" {
					queue = append(queue, d.depTarget)
					continue
				}
				// Non-process dep: aggregate
				if _, ok := entityToProcesses[d.depName]; !ok {
					entityToProcesses[d.depName] = make(map[string]bool)
				}
				entityToProcesses[d.depName][procID] = true
			}
		}
	}

	// AssetX should be linked to both procA (transitively) and procB (directly)
	assetXProcs, ok := entityToProcesses["AssetX"]
	if !ok {
		t.Fatal("AssetX not found in dependency aggregation")
	}
	if !assetXProcs["procA"] {
		t.Error("AssetX should be transitively linked to procA")
	}
	if !assetXProcs["procB"] {
		t.Error("AssetX should be directly linked to procB")
	}
	if len(assetXProcs) != 2 {
		t.Errorf("Expected AssetX linked to 2 critical processes, got %d", len(assetXProcs))
	}
}

// TestSPOFNoFalsePositives ensures non-critical processes don't contribute to SPoF counts.
func TestSPOFNoFalsePositives(t *testing.T) {
	// Only 1 critical process depends on each entity.
	// 5 non-critical processes also depend on the same entity.
	// Should NOT be flagged as SPoF (only 1 critical process).

	critCount := 1
	nonCritCount := 5

	type procDep struct {
		isCritical bool
		depName    string
	}

	var deps []procDep
	for i := 0; i < critCount; i++ {
		deps = append(deps, procDep{isCritical: true, depName: "SharedDB"})
	}
	for i := 0; i < nonCritCount; i++ {
		deps = append(deps, procDep{isCritical: false, depName: "SharedDB"})
	}

	critProcessDeps := make(map[string]int)
	for _, d := range deps {
		if d.isCritical {
			critProcessDeps[d.depName]++
		}
	}

	for name, count := range critProcessDeps {
		if count >= 2 {
			t.Errorf("Entity %s should not be a SPoF (only %d critical process), but threshold met with count %d",
				name, critCount, count)
		}
	}
}

// ============================================================
// Financial Impact Aggregation Tests
// ============================================================

func TestFinancialImpactAggregation(t *testing.T) {
	type process struct {
		hourlyImpact *float64
		dailyImpact  *float64
	}

	h1 := 500.0
	h2 := 1200.0
	d1 := 12000.0
	d2 := 28800.0

	processes := []process{
		{hourlyImpact: &h1, dailyImpact: &d1},
		{hourlyImpact: &h2, dailyImpact: &d2},
		{hourlyImpact: nil, dailyImpact: nil}, // process without financial assessment
	}

	var totalHourly, totalDaily float64
	for _, p := range processes {
		if p.hourlyImpact != nil {
			totalHourly += *p.hourlyImpact
		}
		if p.dailyImpact != nil {
			totalDaily += *p.dailyImpact
		}
	}

	expectedHourly := 1700.0
	expectedDaily := 40800.0
	expectedWeekly := expectedDaily * 5

	if totalHourly != expectedHourly {
		t.Errorf("Total hourly impact = %.2f, want %.2f", totalHourly, expectedHourly)
	}
	if totalDaily != expectedDaily {
		t.Errorf("Total daily impact = %.2f, want %.2f", totalDaily, expectedDaily)
	}
	weeklyImpact := totalDaily * 5
	if weeklyImpact != expectedWeekly {
		t.Errorf("Total weekly impact = %.2f, want %.2f", weeklyImpact, expectedWeekly)
	}
}

func TestFinancialImpactAggregation_AllNil(t *testing.T) {
	type process struct {
		hourlyImpact *float64
		dailyImpact  *float64
	}

	processes := []process{
		{hourlyImpact: nil, dailyImpact: nil},
		{hourlyImpact: nil, dailyImpact: nil},
	}

	var totalHourly, totalDaily float64
	for _, p := range processes {
		if p.hourlyImpact != nil {
			totalHourly += *p.hourlyImpact
		}
		if p.dailyImpact != nil {
			totalDaily += *p.dailyImpact
		}
	}

	if totalHourly != 0 {
		t.Errorf("Expected 0 hourly impact when all nil, got %.2f", totalHourly)
	}
	if totalDaily != 0 {
		t.Errorf("Expected 0 daily impact when all nil, got %.2f", totalDaily)
	}
}

func TestFinancialImpactAggregation_SingleProcess(t *testing.T) {
	hourly := 250.0
	daily := 6000.0

	type process struct {
		hourlyImpact *float64
		dailyImpact  *float64
	}

	processes := []process{
		{hourlyImpact: &hourly, dailyImpact: &daily},
	}

	var totalHourly, totalDaily float64
	for _, p := range processes {
		if p.hourlyImpact != nil {
			totalHourly += *p.hourlyImpact
		}
		if p.dailyImpact != nil {
			totalDaily += *p.dailyImpact
		}
	}

	if totalHourly != 250.0 {
		t.Errorf("Expected 250.00, got %.2f", totalHourly)
	}
	if totalDaily != 6000.0 {
		t.Errorf("Expected 6000.00, got %.2f", totalDaily)
	}
}

// ============================================================
// RTO/RPO Summary Statistics Tests
// ============================================================

func TestRTORPOStatistics(t *testing.T) {
	rto1 := 4.0
	rto2 := 8.0
	rto3 := 24.0
	rpo1 := 1.0
	rpo2 := 4.0

	type process struct {
		rto *float64
		rpo *float64
	}

	processes := []process{
		{rto: &rto1, rpo: &rpo1},
		{rto: &rto2, rpo: &rpo2},
		{rto: &rto3, rpo: nil}, // process without RPO
		{rto: nil, rpo: nil},   // process without either
	}

	var rtoCount, rpoCount int
	var rtoMin, rtoMax, rtoSum float64
	var rpoMin, rpoMax, rpoSum float64
	rtoFirst := true
	rpoFirst := true

	for _, p := range processes {
		if p.rto != nil {
			v := *p.rto
			rtoCount++
			rtoSum += v
			if rtoFirst || v < rtoMin {
				rtoMin = v
			}
			if rtoFirst || v > rtoMax {
				rtoMax = v
			}
			rtoFirst = false
		}
		if p.rpo != nil {
			v := *p.rpo
			rpoCount++
			rpoSum += v
			if rpoFirst || v < rpoMin {
				rpoMin = v
			}
			if rpoFirst || v > rpoMax {
				rpoMax = v
			}
			rpoFirst = false
		}
	}

	if rtoCount != 3 {
		t.Errorf("RTO count = %d, want 3", rtoCount)
	}
	if rpoCount != 2 {
		t.Errorf("RPO count = %d, want 2", rpoCount)
	}
	if rtoMin != 4.0 {
		t.Errorf("Min RTO = %.1f, want 4.0", rtoMin)
	}
	if rtoMax != 24.0 {
		t.Errorf("Max RTO = %.1f, want 24.0", rtoMax)
	}
	expectedRTOAvg := (4.0 + 8.0 + 24.0) / 3
	rtoAvg := rtoSum / float64(rtoCount)
	if rtoAvg != expectedRTOAvg {
		t.Errorf("Avg RTO = %.2f, want %.2f", rtoAvg, expectedRTOAvg)
	}
	if rpoMin != 1.0 {
		t.Errorf("Min RPO = %.1f, want 1.0", rpoMin)
	}
	if rpoMax != 4.0 {
		t.Errorf("Max RPO = %.1f, want 4.0", rpoMax)
	}
}
