package top.hcode.hoj.controller.oj;

import org.apache.shiro.authz.annotation.Logical;
import org.apache.shiro.authz.annotation.RequiresAuthentication;
import org.apache.shiro.authz.annotation.RequiresRoles;
import org.springframework.web.bind.annotation.*;
import top.hcode.hoj.common.result.CommonResult;
import top.hcode.hoj.pojo.dto.*;
import top.hcode.hoj.pojo.entity.learning.LearningMap;
import top.hcode.hoj.pojo.entity.learning.LearningMapEdge;
import top.hcode.hoj.pojo.entity.learning.LearningMapNode;
import top.hcode.hoj.pojo.vo.*;
import top.hcode.hoj.service.learning.LearningMapService;

import javax.annotation.Resource;
import java.util.List;

@RestController
@RequestMapping("/api")
@RequiresAuthentication
public class LearningMapController {
    @Resource
    private LearningMapService service;

    @GetMapping("/learning-maps")
    public CommonResult<List<LearningMap>> publishedMaps() { return service.publishedMaps(); }

    @GetMapping("/learning-maps/{mapId}")
    public CommonResult<LearningMapGraphVO> graph(@PathVariable Long mapId) { return service.graph(mapId); }

    @GetMapping("/learning-maps/{mapId}/progress")
    public CommonResult<LearningMapProgressVO> progress(@PathVariable Long mapId) { return service.progress(mapId); }

    @GetMapping("/learning-maps/{mapId}/full")
    public CommonResult<LearningMapFullVO> full(@PathVariable Long mapId) { return service.full(mapId); }

    @PostMapping("/learning-maps/{mapId}/nodes/{nodeId}/start")
    public CommonResult<Void> start(@PathVariable Long mapId, @PathVariable Long nodeId) { return service.start(mapId, nodeId); }

    @PostMapping("/learning-maps/{mapId}/nodes/{nodeId}/complete")
    public CommonResult<Void> complete(@PathVariable Long mapId, @PathVariable Long nodeId) { return service.complete(mapId, nodeId); }

    @GetMapping("/learning-maps/{mapId}/search")
    public CommonResult<List<LearningMapNodeVO>> search(@PathVariable Long mapId, @RequestParam(required = false) String q) {
        return service.search(mapId, q);
    }

    @GetMapping("/learning-maps/{mapId}/recommend-next")
    public CommonResult<LearningMapRecommendVO> recommend(@PathVariable Long mapId) { return service.recommend(mapId); }

    @GetMapping("/admin/learning-maps")
    @RequiresRoles(value = {"root", "admin", "problem_admin"}, logical = Logical.OR)
    public CommonResult<List<LearningMap>> adminList() { return service.adminList(); }

    @GetMapping("/admin/learning-maps/{mapId}")
    @RequiresRoles(value = {"root", "admin", "problem_admin"}, logical = Logical.OR)
    public CommonResult<LearningMapGraphVO> adminGraph(@PathVariable Long mapId) { return service.adminGraph(mapId); }

    @PostMapping("/admin/learning-maps")
    @RequiresRoles(value = {"root", "admin", "problem_admin"}, logical = Logical.OR)
    public CommonResult<LearningMap> create(@RequestBody LearningMapDTO dto) { return service.create(dto); }

    @PutMapping("/admin/learning-maps/{mapId}")
    @RequiresRoles(value = {"root", "admin", "problem_admin"}, logical = Logical.OR)
    public CommonResult<LearningMap> update(@PathVariable Long mapId, @RequestBody LearningMapDTO dto) { return service.update(mapId, dto); }

    @DeleteMapping("/admin/learning-maps/{mapId}")
    @RequiresRoles(value = {"root", "admin", "problem_admin"}, logical = Logical.OR)
    public CommonResult<Void> delete(@PathVariable Long mapId) { return service.delete(mapId); }

    @PostMapping("/admin/learning-maps/{mapId}/nodes")
    @RequiresRoles(value = {"root", "admin", "problem_admin"}, logical = Logical.OR)
    public CommonResult<LearningMapNode> createNode(@PathVariable Long mapId, @RequestBody LearningMapNodeDTO dto) { return service.createNode(mapId, dto); }

    @PutMapping("/admin/learning-maps/{mapId}/nodes/{nodeId}")
    @RequiresRoles(value = {"root", "admin", "problem_admin"}, logical = Logical.OR)
    public CommonResult<LearningMapNode> updateNode(@PathVariable Long mapId, @PathVariable Long nodeId, @RequestBody LearningMapNodeDTO dto) { return service.updateNode(mapId, nodeId, dto); }

    @DeleteMapping("/admin/learning-maps/{mapId}/nodes/{nodeId}")
    @RequiresRoles(value = {"root", "admin", "problem_admin"}, logical = Logical.OR)
    public CommonResult<Void> deleteNode(@PathVariable Long mapId, @PathVariable Long nodeId) { return service.deleteNode(mapId, nodeId); }

    @PostMapping("/admin/learning-maps/{mapId}/edges")
    @RequiresRoles(value = {"root", "admin", "problem_admin"}, logical = Logical.OR)
    public CommonResult<LearningMapEdge> createEdge(@PathVariable Long mapId, @RequestBody LearningMapEdgeDTO dto) { return service.createEdge(mapId, dto); }

    @PutMapping("/admin/learning-maps/{mapId}/edges/{edgeId}")
    @RequiresRoles(value = {"root", "admin", "problem_admin"}, logical = Logical.OR)
    public CommonResult<LearningMapEdge> updateEdge(@PathVariable Long mapId, @PathVariable Long edgeId, @RequestBody LearningMapEdgeDTO dto) { return service.updateEdge(mapId, edgeId, dto); }

    @DeleteMapping("/admin/learning-maps/{mapId}/edges/{edgeId}")
    @RequiresRoles(value = {"root", "admin", "problem_admin"}, logical = Logical.OR)
    public CommonResult<Void> deleteEdge(@PathVariable Long mapId, @PathVariable Long edgeId) { return service.deleteEdge(mapId, edgeId); }

    @PostMapping("/admin/learning-maps/{mapId}/publish")
    @RequiresRoles(value = {"root", "admin", "problem_admin"}, logical = Logical.OR)
    public CommonResult<Void> publish(@PathVariable Long mapId) { return service.publish(mapId); }

    @GetMapping("/admin/learning-maps/{mapId}/validate")
    @RequiresRoles(value = {"root", "admin", "problem_admin"}, logical = Logical.OR)
    public CommonResult<LearningMapValidationVO> validate(@PathVariable Long mapId) { return service.validate(mapId); }

    @GetMapping("/admin/problems/search")
    @RequiresRoles(value = {"root", "admin", "problem_admin"}, logical = Logical.OR)
    public CommonResult<List<LearningProblemVO>> searchProblems(@RequestParam(required = false) String q) { return service.searchProblems(q); }

    @GetMapping("/admin/problems/{problemId}")
    @RequiresRoles(value = {"root", "admin", "problem_admin"}, logical = Logical.OR)
    public CommonResult<LearningProblemVO> getProblem(@PathVariable String problemId) { return service.getProblem(problemId); }

    @GetMapping("/admin/learning-maps/{mapId}/permissions")
    @RequiresRoles(value = {"root", "admin", "problem_admin"}, logical = Logical.OR)
    public CommonResult<LearningMapAccessVO> access(@PathVariable Long mapId) { return service.access(mapId); }

    @PutMapping("/admin/learning-maps/{mapId}/permissions/mode")
    @RequiresRoles(value = {"root", "admin", "problem_admin"}, logical = Logical.OR)
    public CommonResult<Void> setMode(@PathVariable Long mapId, @RequestBody LearningMapAccessModeDTO dto) { return service.setAccessMode(mapId, dto); }

    @PutMapping("/admin/learning-maps/{mapId}/permissions/{userId}")
    @RequiresRoles(value = {"root", "admin", "problem_admin"}, logical = Logical.OR)
    public CommonResult<Void> setPermission(@PathVariable Long mapId, @PathVariable String userId, @RequestBody LearningMapPermissionDTO dto) { return service.setPermission(mapId, userId, dto); }

    @PostMapping("/admin/learning-maps/{mapId}/permissions/batch")
    @RequiresRoles(value = {"root", "admin", "problem_admin"}, logical = Logical.OR)
    public CommonResult<Void> batchPermission(@PathVariable Long mapId, @RequestBody LearningMapBatchPermissionDTO dto) { return service.batchPermission(mapId, dto); }

    @DeleteMapping("/admin/learning-maps/{mapId}/permissions/{userId}")
    @RequiresRoles(value = {"root", "admin", "problem_admin"}, logical = Logical.OR)
    public CommonResult<Void> deletePermission(@PathVariable Long mapId, @PathVariable String userId) { return service.deletePermission(mapId, userId); }

    @GetMapping("/admin/learning-maps/users/search")
    @RequiresRoles(value = {"root", "admin", "problem_admin"}, logical = Logical.OR)
    public CommonResult<List<LearningMapPermissionUserVO>> searchUsers(@RequestParam(required = false) String q) { return service.searchUsers(q); }
}
