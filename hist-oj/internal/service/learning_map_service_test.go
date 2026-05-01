package service

import (
	"errors"
	"testing"

	"github.com/hoj/hist-oj/internal/model"
)

func u64Ptr(v uint64) *uint64 { return &v }

func TestLockedNodeCannotComplete(t *testing.T) {
	err := ValidateKnowledgeActionAllowed(model.LearningMapNodeTypeKnowledge, model.LearningMapProgressLocked)
	if !errors.Is(err, ErrLearningNodeLocked) {
		t.Fatalf("expected ErrLearningNodeLocked, got %v", err)
	}
}

func TestPrerequisiteCompletedNodeBecomesAvailable(t *testing.T) {
	nodes := []model.LearningMapNode{
		{ID: 1, Type: model.LearningMapNodeTypeKnowledge, Title: "A"},
		{ID: 2, Type: model.LearningMapNodeTypeKnowledge, Title: "B"},
	}
	edges := []model.LearningMapEdge{
		{ID: 1, SourceNodeID: 1, TargetNodeID: 2, Type: model.LearningMapEdgeTypePrerequisite},
	}
	existing := map[uint64]model.UserLearningProgress{
		1: {NodeID: 1, Status: model.LearningMapProgressCompleted},
	}
	snapshot := ComputeLearningNodeProgress(nodes, edges, existing, map[uint64]bool{})
	if snapshot[2].Status != model.LearningMapProgressAvailable {
		t.Fatalf("expected node 2 available, got %s", snapshot[2].Status)
	}
}

func TestProblemNodeMustBeCompletedByAC(t *testing.T) {
	nodes := []model.LearningMapNode{
		{ID: 1, Type: model.LearningMapNodeTypeProblem, ProblemID: u64Ptr(1001), Title: "P1"},
	}
	existing := map[uint64]model.UserLearningProgress{
		1: {NodeID: 1, Status: model.LearningMapProgressCompleted},
	}

	// 未 AC：即便存在手动 completed 记录，也不能判定为 completed。
	snapshotNoAC := ComputeLearningNodeProgress(nodes, nil, existing, map[uint64]bool{})
	if snapshotNoAC[1].Status == model.LearningMapProgressCompleted || snapshotNoAC[1].Status == model.LearningMapProgressMastered {
		t.Fatalf("problem node should not be completed without AC, got %s", snapshotNoAC[1].Status)
	}

	// AC 后自动变 completed。
	snapshotAC := ComputeLearningNodeProgress(nodes, nil, existing, map[uint64]bool{1001: true})
	if snapshotAC[1].Status != model.LearningMapProgressCompleted {
		t.Fatalf("expected completed after AC, got %s", snapshotAC[1].Status)
	}
}

func TestLockedPrerequisiteBlocksDownstreamNodeEvenIfSolved(t *testing.T) {
	nodes := []model.LearningMapNode{
		{ID: 1, Type: model.LearningMapNodeTypeKnowledge, Title: "A"},
		{ID: 2, Type: model.LearningMapNodeTypeProblem, ProblemID: u64Ptr(1001), Title: "B"},
		{ID: 3, Type: model.LearningMapNodeTypeKnowledge, Title: "C"},
	}
	edges := []model.LearningMapEdge{
		{ID: 1, SourceNodeID: 1, TargetNodeID: 2, Type: model.LearningMapEdgeTypePrerequisite},
		{ID: 2, SourceNodeID: 2, TargetNodeID: 3, Type: model.LearningMapEdgeTypePrerequisite},
	}
	solved := map[uint64]bool{1001: true}

	lockedSnapshot := ComputeLearningNodeProgress(
		nodes,
		edges,
		map[uint64]model.UserLearningProgress{
			1: {NodeID: 1, Status: model.LearningMapProgressInProgress},
		},
		solved,
	)
	if lockedSnapshot[2].Status != model.LearningMapProgressLocked {
		t.Fatalf("expected node B locked when A not completed, got %s", lockedSnapshot[2].Status)
	}
	if lockedSnapshot[3].Status != model.LearningMapProgressLocked {
		t.Fatalf("expected node C locked when B is effectively locked, got %s", lockedSnapshot[3].Status)
	}

	unlockedSnapshot := ComputeLearningNodeProgress(
		nodes,
		edges,
		map[uint64]model.UserLearningProgress{
			1: {NodeID: 1, Status: model.LearningMapProgressCompleted},
		},
		solved,
	)
	if unlockedSnapshot[2].Status != model.LearningMapProgressCompleted {
		t.Fatalf("expected node B completed after A completed, got %s", unlockedSnapshot[2].Status)
	}
	if unlockedSnapshot[3].Status != model.LearningMapProgressAvailable {
		t.Fatalf("expected node C available after B completed, got %s", unlockedSnapshot[3].Status)
	}
}

func TestKnowledgeNodeCanBeManuallyCompleted(t *testing.T) {
	nodes := []model.LearningMapNode{
		{ID: 1, Type: model.LearningMapNodeTypeKnowledge, Title: "K1"},
	}
	existing := map[uint64]model.UserLearningProgress{
		1: {NodeID: 1, Status: model.LearningMapProgressCompleted},
	}
	snapshot := ComputeLearningNodeProgress(nodes, nil, existing, map[uint64]bool{})
	if snapshot[1].Status != model.LearningMapProgressCompleted {
		t.Fatalf("expected knowledge node completed, got %s", snapshot[1].Status)
	}
}

func TestDetectPrerequisiteCycle(t *testing.T) {
	nodes := []model.LearningMapNode{
		{ID: 1, Type: model.LearningMapNodeTypeKnowledge},
		{ID: 2, Type: model.LearningMapNodeTypeKnowledge},
	}
	edges := []model.LearningMapEdge{
		{SourceNodeID: 1, TargetNodeID: 2, Type: model.LearningMapEdgeTypePrerequisite},
		{SourceNodeID: 2, TargetNodeID: 1, Type: model.LearningMapEdgeTypePrerequisite},
	}
	hasCycle, cycleNodes := DetectPrerequisiteCycle(nodes, edges)
	if !hasCycle {
		t.Fatalf("expected cycle, got none")
	}
	if len(cycleNodes) == 0 {
		t.Fatalf("expected cycle nodes, got empty")
	}
}

func TestCleanupEdgesAfterNodeDelete(t *testing.T) {
	edges := []model.LearningMapEdge{
		{ID: 1, SourceNodeID: 1, TargetNodeID: 2},
		{ID: 2, SourceNodeID: 2, TargetNodeID: 3},
		{ID: 3, SourceNodeID: 4, TargetNodeID: 5},
	}
	cleaned := CleanupEdgesAfterNodeDelete(edges, 2)
	if len(cleaned) != 1 {
		t.Fatalf("expected 1 edge after cleanup, got %d", len(cleaned))
	}
	if cleaned[0].ID != 3 {
		t.Fatalf("expected edge 3 to remain, got edge %d", cleaned[0].ID)
	}
}

func TestCheckPrerequisiteEdgeCreatesCycle(t *testing.T) {
	edges := []model.LearningMapEdge{
		{ID: 1, SourceNodeID: 1, TargetNodeID: 2, Type: model.LearningMapEdgeTypePrerequisite},
		{ID: 2, SourceNodeID: 2, TargetNodeID: 3, Type: model.LearningMapEdgeTypePrerequisite},
	}
	hasCycle, cyclePath := CheckPrerequisiteEdgeCreatesCycle(edges, 0, 3, 1)
	if !hasCycle {
		t.Fatalf("expected cycle on adding 3->1")
	}
	if len(cyclePath) != 4 {
		t.Fatalf("expected cycle path length=4, got %d", len(cyclePath))
	}
	expected := []uint64{3, 1, 2, 3}
	for i := range expected {
		if cyclePath[i] != expected[i] {
			t.Fatalf("unexpected cycle path at idx=%d, got=%v expected=%v", i, cyclePath, expected)
		}
	}
}

func TestCheckPrerequisiteEdgeCreatesCycleWithIgnoreEdgeID(t *testing.T) {
	// 更新 edge#10 时，旧边 2->1 应被忽略；否则会误判新边 3->2 成环。
	edges := []model.LearningMapEdge{
		{ID: 10, SourceNodeID: 2, TargetNodeID: 1, Type: model.LearningMapEdgeTypePrerequisite},
		{ID: 11, SourceNodeID: 1, TargetNodeID: 3, Type: model.LearningMapEdgeTypePrerequisite},
	}
	hasCycle, cyclePath := CheckPrerequisiteEdgeCreatesCycle(edges, 10, 3, 2)
	if hasCycle {
		t.Fatalf("expected no cycle when ignoring old edge, got cycle path=%v", cyclePath)
	}
}

func TestMapAccessDefaultAllOpen(t *testing.T) {
	if !isMapAccessAllowed(model.LearningMapAccessAllOpen, false, false) {
		t.Fatalf("expected all_open map to allow access by default")
	}
}

func TestMapAccessAllClosedWithUserOverride(t *testing.T) {
	if isMapAccessAllowed(model.LearningMapAccessAllClosed, false, false) {
		t.Fatalf("expected all_closed map to deny access by default")
	}
	if !isMapAccessAllowed(model.LearningMapAccessAllClosed, true, true) {
		t.Fatalf("expected explicit enabled override to allow access")
	}
	if isMapAccessAllowed(model.LearningMapAccessAllOpen, true, false) {
		t.Fatalf("expected explicit disabled override to deny access")
	}
}
