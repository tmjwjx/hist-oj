package top.hcode.hoj.service.learning;

import top.hcode.hoj.common.result.CommonResult;
import top.hcode.hoj.pojo.dto.*;
import top.hcode.hoj.pojo.entity.learning.LearningMap;
import top.hcode.hoj.pojo.entity.learning.LearningMapEdge;
import top.hcode.hoj.pojo.entity.learning.LearningMapNode;
import top.hcode.hoj.pojo.vo.*;

import java.util.List;

public interface LearningMapService {
    CommonResult<List<LearningMap>> publishedMaps();
    CommonResult<LearningMapGraphVO> graph(Long mapId);
    CommonResult<LearningMapProgressVO> progress(Long mapId);
    CommonResult<LearningMapFullVO> full(Long mapId);
    CommonResult<Void> start(Long mapId, Long nodeId);
    CommonResult<Void> complete(Long mapId, Long nodeId);
    CommonResult<List<LearningMapNodeVO>> search(Long mapId, String keyword);
    CommonResult<LearningMapRecommendVO> recommend(Long mapId);
    CommonResult<List<LearningMap>> adminList();
    CommonResult<LearningMapGraphVO> adminGraph(Long mapId);
    CommonResult<LearningMap> create(LearningMapDTO dto);
    CommonResult<LearningMap> update(Long mapId, LearningMapDTO dto);
    CommonResult<Void> delete(Long mapId);
    CommonResult<LearningMapNode> createNode(Long mapId, LearningMapNodeDTO dto);
    CommonResult<LearningMapNode> updateNode(Long mapId, Long nodeId, LearningMapNodeDTO dto);
    CommonResult<Void> deleteNode(Long mapId, Long nodeId);
    CommonResult<LearningMapEdge> createEdge(Long mapId, LearningMapEdgeDTO dto);
    CommonResult<LearningMapEdge> updateEdge(Long mapId, Long edgeId, LearningMapEdgeDTO dto);
    CommonResult<Void> deleteEdge(Long mapId, Long edgeId);
    CommonResult<Void> publish(Long mapId);
    CommonResult<LearningMapValidationVO> validate(Long mapId);
    CommonResult<List<LearningProblemVO>> searchProblems(String keyword);
    CommonResult<LearningProblemVO> getProblem(String identifier);
    CommonResult<LearningMapAccessVO> access(Long mapId);
    CommonResult<Void> setAccessMode(Long mapId, LearningMapAccessModeDTO dto);
    CommonResult<Void> setPermission(Long mapId, String userId, LearningMapPermissionDTO dto);
    CommonResult<Void> batchPermission(Long mapId, LearningMapBatchPermissionDTO dto);
    CommonResult<Void> deletePermission(Long mapId, String userId);
    CommonResult<List<LearningMapPermissionUserVO>> searchUsers(String keyword);
}
