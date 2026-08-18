package top.hcode.hoj.service.learning.impl;

import lombok.extern.slf4j.Slf4j;
import org.springframework.stereotype.Service;
import top.hcode.hoj.common.exception.LearningMapException;
import top.hcode.hoj.common.result.CommonResult;
import top.hcode.hoj.common.result.ResultStatus;
import top.hcode.hoj.manager.learning.LearningMapCatalogManager;
import top.hcode.hoj.manager.learning.LearningMapPermissionManager;
import top.hcode.hoj.manager.learning.LearningMapProblemManager;
import top.hcode.hoj.manager.learning.LearningMapProgressManager;
import top.hcode.hoj.pojo.dto.*;
import top.hcode.hoj.pojo.entity.learning.LearningMap;
import top.hcode.hoj.pojo.entity.learning.LearningMapEdge;
import top.hcode.hoj.pojo.entity.learning.LearningMapNode;
import top.hcode.hoj.pojo.vo.*;
import top.hcode.hoj.service.learning.LearningMapService;
import top.hcode.hoj.utils.ShiroUtils;

import javax.annotation.Resource;
import java.util.List;
import java.util.function.Supplier;

@Slf4j
@Service
public class LearningMapServiceImpl implements LearningMapService {
    @Resource private LearningMapCatalogManager catalog;
    @Resource private LearningMapPermissionManager permissions;
    @Resource private LearningMapProblemManager problems;
    @Resource private LearningMapProgressManager progress;

    @Override public CommonResult<List<LearningMap>> publishedMaps() {
        return run(() -> permissions.listAccessible(uid()));
    }
    @Override public CommonResult<LearningMapGraphVO> graph(Long id) { return run(() -> progress.graph(id, uid())); }
    @Override public CommonResult<LearningMapProgressVO> progress(Long id) { return run(() -> progress.progress(id, uid())); }
    @Override public CommonResult<LearningMapFullVO> full(Long id) { return run(() -> progress.full(id, uid())); }
    @Override public CommonResult<Void> start(Long mapId, Long nodeId) { return runVoid(() -> progress.start(mapId, nodeId, uid())); }
    @Override public CommonResult<Void> complete(Long mapId, Long nodeId) { return runVoid(() -> progress.complete(mapId, nodeId, uid())); }
    @Override public CommonResult<List<LearningMapNodeVO>> search(Long id, String keyword) { return run(() -> progress.search(id, uid(), keyword)); }
    @Override public CommonResult<LearningMapRecommendVO> recommend(Long id) { return run(() -> progress.recommend(id, uid())); }
    @Override public CommonResult<List<LearningMap>> adminList() { return run(() -> catalog.list(false)); }
    @Override public CommonResult<LearningMapGraphVO> adminGraph(Long id) { return run(() -> catalog.graph(id, false)); }
    @Override public CommonResult<LearningMap> create(LearningMapDTO dto) { return run(() -> catalog.create(dto)); }
    @Override public CommonResult<LearningMap> update(Long id, LearningMapDTO dto) { return run(() -> catalog.update(id, dto)); }
    @Override public CommonResult<Void> delete(Long id) { return runVoid(() -> catalog.delete(id)); }
    @Override public CommonResult<LearningMapNode> createNode(Long id, LearningMapNodeDTO dto) { return run(() -> catalog.createNode(id, dto)); }
    @Override public CommonResult<LearningMapNode> updateNode(Long mapId, Long nodeId, LearningMapNodeDTO dto) { return run(() -> catalog.updateNode(mapId, nodeId, dto)); }
    @Override public CommonResult<Void> deleteNode(Long mapId, Long nodeId) { return runVoid(() -> catalog.deleteNode(mapId, nodeId)); }
    @Override public CommonResult<LearningMapEdge> createEdge(Long id, LearningMapEdgeDTO dto) { return run(() -> catalog.createEdge(id, dto)); }
    @Override public CommonResult<LearningMapEdge> updateEdge(Long mapId, Long edgeId, LearningMapEdgeDTO dto) { return run(() -> catalog.updateEdge(mapId, edgeId, dto)); }
    @Override public CommonResult<Void> deleteEdge(Long mapId, Long edgeId) { return runVoid(() -> catalog.deleteEdge(mapId, edgeId)); }
    @Override public CommonResult<Void> publish(Long id) { return runVoid(() -> catalog.publish(id)); }
    @Override public CommonResult<LearningMapValidationVO> validate(Long id) {
        return run(() -> {
            List<top.hcode.hoj.pojo.vo.LearningMapValidationErrorVO> errors = catalog.validate(id);
            return new LearningMapValidationVO().setValid(errors.isEmpty()).setErrors(errors);
        });
    }
    @Override public CommonResult<List<LearningProblemVO>> searchProblems(String keyword) { return run(() -> problems.search(keyword)); }
    @Override public CommonResult<LearningProblemVO> getProblem(String identifier) { return run(() -> problems.get(identifier)); }
    @Override public CommonResult<LearningMapAccessVO> access(Long id) { return run(() -> permissions.config(id)); }
    @Override public CommonResult<Void> setAccessMode(Long id, LearningMapAccessModeDTO dto) { return runVoid(() -> permissions.setAccessMode(id, dto.getAccessMode())); }
    @Override public CommonResult<Void> setPermission(Long id, String userId, LearningMapPermissionDTO dto) { return runVoid(() -> permissions.setPermission(id, userId, Boolean.TRUE.equals(dto.getEnabled()))); }
    @Override public CommonResult<Void> batchPermission(Long id, LearningMapBatchPermissionDTO dto) { return runVoid(() -> permissions.setPermissions(id, dto.getUserIds(), Boolean.TRUE.equals(dto.getEnabled()))); }
    @Override public CommonResult<Void> deletePermission(Long id, String userId) { return runVoid(() -> permissions.deletePermission(id, userId)); }
    @Override public CommonResult<List<LearningMapPermissionUserVO>> searchUsers(String keyword) { return run(() -> permissions.searchUsers(keyword)); }

    private String uid() { return ShiroUtils.getProfile().getUid(); }

    private <T> CommonResult<T> run(Supplier<T> action) {
        try {
            return CommonResult.successResponse(action.get());
        } catch (LearningMapException ex) {
            return CommonResult.errorResponse(ex.getMessage(), ex.getStatus());
        } catch (Exception ex) {
            log.error("learning map request failed", ex);
            return CommonResult.errorResponse("航海图操作失败", ResultStatus.SYSTEM_ERROR);
        }
    }

    private CommonResult<Void> runVoid(Runnable action) {
        return run(() -> { action.run(); return null; });
    }
}
