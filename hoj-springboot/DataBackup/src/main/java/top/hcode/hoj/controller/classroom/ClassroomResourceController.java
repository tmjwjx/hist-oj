package top.hcode.hoj.controller.classroom;

import org.apache.shiro.authz.annotation.RequiresAuthentication;
import org.springframework.core.io.FileSystemResource;
import org.springframework.http.*;
import org.springframework.web.bind.annotation.*;
import org.springframework.web.multipart.MultipartFile;
import top.hcode.hoj.common.result.CommonResult;
import top.hcode.hoj.manager.classroom.ClassroomResourceManager;
import top.hcode.hoj.pojo.entity.classroom.*;

import javax.annotation.Resource;
import java.io.File;
import java.util.*;

@RestController
@RequestMapping("/api/classroom")
public class ClassroomResourceController extends ClassroomControllerSupport {
    @Resource private ClassroomResourceManager manager;

    @PostMapping("/folder") @RequiresAuthentication
    public CommonResult<ClassroomFolder> createFolder(@RequestBody Map<String, Object> request) { return run(() -> manager.createFolder(request)); }
    @GetMapping("/{classroomId}/folders") @RequiresAuthentication
    public CommonResult<List<ClassroomFolder>> folders(@PathVariable Long classroomId, @RequestParam Map<String, String> params) { return run(() -> manager.folders(classroomId, params)); }
    @PutMapping("/folder") @RequiresAuthentication
    public CommonResult<Void> updateFolder(@RequestBody Map<String, Object> request) { return runVoid(() -> manager.updateFolder(request)); }
    @DeleteMapping("/folder/{folderId}") @RequiresAuthentication
    public CommonResult<Void> deleteFolder(@PathVariable Long folderId) { return runVoid(() -> manager.deleteFolder(folderId)); }

    @PostMapping("/material/upload") @RequiresAuthentication
    public CommonResult<ClassroomMaterial> upload(@RequestParam Long classroomId, @RequestParam(required = false, defaultValue = "0") Long folderId, @RequestParam(required = false, defaultValue = "0") Integer isShared, @RequestParam("file") MultipartFile file) { return run(() -> manager.upload(classroomId, folderId, isShared, file)); }
    @GetMapping("/{classroomId}/folder/{folderId}/materials") @RequiresAuthentication
    public CommonResult<List<Map<String, Object>>> materials(@PathVariable Long classroomId, @PathVariable String folderId) { return run(() -> manager.materials(classroomId, "root".equals(folderId) ? 0L : Long.valueOf(folderId))); }
    @DeleteMapping("/material/{materialId}") @RequiresAuthentication
    public CommonResult<Void> deleteMaterial(@PathVariable Long materialId) { return runVoid(() -> manager.deleteMaterial(materialId)); }
    @PostMapping("/material/copy") @RequiresAuthentication
    public CommonResult<ClassroomMaterial> copy(@RequestBody Map<String, Object> request) { return run(() -> manager.copy(request)); }

    @GetMapping("/material/{materialId}/permissions") @RequiresAuthentication
    public CommonResult<List<Map<String, Object>>> permissions(@PathVariable Long materialId) { return run(() -> manager.permissions(materialId)); }
    @PostMapping("/material/permissions") @RequiresAuthentication
    public CommonResult<Void> permissions(@RequestBody Map<String, Object> request) { return runVoid(() -> manager.setPermissions(number(request.get("materialId")), list(request.get("permissions")))); }
    @PostMapping("/material/{materialId}/permissions/batch") @RequiresAuthentication
    public CommonResult<Void> permissionsBatch(@PathVariable Long materialId, @RequestBody Map<String, Object> request) { return runVoid(() -> manager.setAllPermissions(materialId, bool(request.get("canPreview")), bool(request.get("canDownload")))); }

    @GetMapping("/material/{materialId}/download") @RequiresAuthentication
    public ResponseEntity<?> download(@PathVariable Long materialId) { return fileResponse(materialId, true); }
    @GetMapping("/material/{materialId}/pdf/binary") @RequiresAuthentication
    public ResponseEntity<?> pdf(@PathVariable Long materialId) { return fileResponse(materialId, false); }
    @GetMapping("/material/{materialId}/pdf") @RequiresAuthentication
    public CommonResult<Map<String, Object>> pdfBase64(@PathVariable Long materialId) { return run(() -> { File file = manager.file(materialId, false); byte[] bytes = java.nio.file.Files.readAllBytes(file.toPath()); Map<String, Object> data = new LinkedHashMap<>(); data.put("fileName", file.getName()); data.put("data", Base64.getEncoder().encodeToString(bytes)); data.put("size", bytes.length); return data; }); }
    @GetMapping("/material/{materialId}/preview-token") @RequiresAuthentication
    public CommonResult<Map<String, Object>> previewToken(@PathVariable Long materialId) { return run(() -> manager.previewToken(materialId)); }
    @GetMapping("/material/preview/{token}")
    public ResponseEntity<?> preview(@PathVariable String token) { try { File file = manager.previewByToken(token); return ResponseEntity.ok().contentType(MediaType.APPLICATION_OCTET_STREAM).header(HttpHeaders.CONTENT_DISPOSITION, "inline; filename=\"" + file.getName() + "\"").body(new FileSystemResource(file)); } catch (Exception e) { return ResponseEntity.status(HttpStatus.FORBIDDEN).body(e.getMessage()); } }
    @GetMapping("/material/{materialId}/cos-preview-url") @RequiresAuthentication
    public CommonResult<Map<String, Object>> cosPreview(@PathVariable Long materialId) { return run(() -> { Map<String, Object> value = new LinkedHashMap<>(); value.put("previewUrl", "/api/classroom/material/" + materialId + "/download"); return value; }); }

    @PostMapping("/{classroomId}/random-pick") @RequiresAuthentication
    public CommonResult<Map<String, Object>> random(@PathVariable Long classroomId) { return run(() -> manager.randomPick(classroomId)); }
    @GetMapping("/{classroomId}/pick-history") @RequiresAuthentication
    public CommonResult<List<Map<String, Object>>> history(@PathVariable Long classroomId, @RequestParam(defaultValue = "1") int page, @RequestParam(defaultValue = "20") int limit) { return run(() -> manager.pickHistory(classroomId, page, limit)); }

    @PostMapping("/message") @RequiresAuthentication
    public CommonResult<Map<String, Object>> message(@RequestBody Map<String, Object> request) { return run(() -> manager.sendMessage(request)); }
    @GetMapping("/{classroomId}/messages") @RequiresAuthentication
    public CommonResult<List<Map<String, Object>>> messages(@PathVariable Long classroomId, @RequestParam(defaultValue = "1") int page, @RequestParam(defaultValue = "50") int limit) { return run(() -> manager.messages(classroomId, page, limit)); }
    @DeleteMapping("/message/{messageId}") @RequiresAuthentication
    public CommonResult<Void> recall(@PathVariable Long messageId) { return runVoid(() -> manager.recallMessage(messageId)); }
    @PostMapping("/message/upload-image") @RequiresAuthentication
    public CommonResult<Map<String, Object>> messageImage(@RequestParam Long classroomId, @RequestParam("file") MultipartFile file) { return run(() -> manager.uploadImage(classroomId, file, "msg_")); }
    @PostMapping("/question/upload-image") @RequiresAuthentication
    public CommonResult<Map<String, Object>> questionImage(@RequestParam(required = false) Long classroomId, @RequestParam("file") MultipartFile file) { return run(() -> manager.uploadImage(classroomId == null ? 0L : classroomId, file, "question_")); }
    @PostMapping("/question/delete-image") @RequiresAuthentication
    public CommonResult<Void> deleteQuestionImage(@RequestBody Map<String, Object> request) { return runVoid(() -> { /* 图片在内容中引用时不自动删除，避免误删共用资源。 */ }); }
    @PostMapping("/homework/upload-attachment") @RequiresAuthentication
    public CommonResult<Map<String, Object>> attachment(@RequestParam Long homeworkId, @RequestParam("file") MultipartFile file) { return run(() -> manager.uploadImage(0L, file, "attachment_")); }

    private ResponseEntity<?> fileResponse(Long id, boolean download) { try { File file = manager.file(id, download); MediaType type = MediaType.APPLICATION_OCTET_STREAM; String name = file.getName().toLowerCase(Locale.ROOT); if (name.endsWith(".pdf")) type = MediaType.APPLICATION_PDF; return ResponseEntity.ok().contentType(type).header(HttpHeaders.CONTENT_DISPOSITION, (download ? "attachment" : "inline") + "; filename=\"" + file.getName() + "\"").body(new FileSystemResource(file)); } catch (Exception e) { return ResponseEntity.status(HttpStatus.FORBIDDEN).body(e.getMessage()); } }
    private Long number(Object value) { try { return value instanceof Number ? ((Number) value).longValue() : Long.parseLong(String.valueOf(value)); } catch (Exception e) { return null; } }
    private boolean bool(Object value) { return value instanceof Boolean ? (Boolean) value : "1".equals(String.valueOf(value)) || "true".equalsIgnoreCase(String.valueOf(value)); }
    @SuppressWarnings("unchecked") private List<Map<String, Object>> list(Object value) { return value instanceof List ? (List<Map<String, Object>>) value : Collections.emptyList(); }
}
