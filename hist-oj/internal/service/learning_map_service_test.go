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
