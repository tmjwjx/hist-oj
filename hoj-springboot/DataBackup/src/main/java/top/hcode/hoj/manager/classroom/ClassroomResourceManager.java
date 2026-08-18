package top.hcode.hoj.manager.classroom;

import com.baomidou.mybatisplus.core.conditions.query.QueryWrapper;
import com.baomidou.mybatisplus.core.conditions.update.UpdateWrapper;
import org.springframework.stereotype.Component;
import org.springframework.transaction.annotation.Transactional;
import org.springframework.web.multipart.MultipartFile;
import top.hcode.hoj.common.exception.StatusFailException;
import top.hcode.hoj.common.exception.StatusForbiddenException;
import top.hcode.hoj.common.exception.StatusNotFoundException;
import top.hcode.hoj.dao.user.UserInfoEntityService;
import top.hcode.hoj.mapper.classroom.*;
import top.hcode.hoj.pojo.entity.classroom.*;
import top.hcode.hoj.pojo.entity.user.UserInfo;
import top.hcode.hoj.pojo.vo.classroom.ClassroomUserVO;

import javax.annotation.Resource;
import java.io.File;
import java.io.IOException;
import java.nio.file.Files;
import java.nio.file.Path;
import java.nio.file.Paths;
import java.security.MessageDigest;
import java.util.*;
import java.util.stream.Collectors;

@Component
public class ClassroomResourceManager {
    @Resource private ClassroomFolderMapper folderMapper;
    @Resource private ClassroomMaterialMapper materialMapper;
    @Resource private ClassroomMaterialPermissionMapper permissionMapper;
    @Resource private ClassroomRandomPickMapper pickMapper;
    @Resource private ClassroomMessageMapper messageMapper;
    @Resource private ClassroomStudentMapper studentMapper;
    @Resource private ClassroomAccessManager access;
    @Resource private UserInfoEntityService userService;

    public ClassroomFolder createFolder(Map<String, Object> request) throws Exception {
        Long classroomId = number(request.get("classroomId")); access.requireManager(classroomId);
        ClassroomFolder row = new ClassroomFolder().setClassroomId(classroomId).setFolderName(text(request.get("folderName"))).setParentId(number(request.get("parentId")) == null ? 0L : number(request.get("parentId"))).setCreatorId(access.uid()).setSortOrder(0).setStatus(1).setCreateTime(new Date()).setUpdateTime(new Date()); folderMapper.insert(row); return row;
    }
    public List<ClassroomFolder> folders(Long classroomId, Map<String, String> params) throws Exception { viewable(classroomId); Long parent = number(params.get("parentId")); if (parent == null) parent = 0L; return folderMapper.selectList(new QueryWrapper<ClassroomFolder>().eq("classroom_id", classroomId).eq("parent_id", parent).eq("status", 1).orderByAsc("sort_order").orderByAsc("create_time")); }
    public void updateFolder(Map<String, Object> request) throws Exception { ClassroomFolder row = folder(number(request.get("id"))); access.requireManager(row.getClassroomId()); folderMapper.update(null, new UpdateWrapper<ClassroomFolder>().eq("id", row.getId()).set("folder_name", text(request.get("folderName"))).set("update_time", new Date())); }
    @Transactional(rollbackFor = Exception.class)
    public void deleteFolder(Long id) throws Exception { ClassroomFolder root = folder(id); access.requireManager(root.getClassroomId()); List<Long> ids = descendants(id); List<ClassroomMaterial> materials = materialMapper.selectList(new QueryWrapper<ClassroomMaterial>().in("folder_id", ids)); for (ClassroomMaterial m : materials) removeFile(m.getFilePath()); if (!ids.isEmpty()) { materialMapper.update(null, new UpdateWrapper<ClassroomMaterial>().in("folder_id", ids).set("status", 0)); folderMapper.update(null, new UpdateWrapper<ClassroomFolder>().in("id", ids).set("status", 0).set("update_time", new Date())); } }

    public ClassroomMaterial upload(Long classroomId, Long folderId, Integer shared, MultipartFile file) throws Exception {
        access.requireManager(classroomId); if (file == null || file.isEmpty()) throw new StatusFailException("请选择要上传的文件");
        if (folderId != null && folderId != 0) { ClassroomFolder folder = folder(folderId); if (!Objects.equals(folder.getClassroomId(), classroomId)) throw new StatusForbiddenException("文件夹不属于当前班级"); }
        String original = file.getOriginalFilename() == null ? "file" : new File(file.getOriginalFilename()).getName(); String stored = System.currentTimeMillis() + "_" + UUID.randomUUID().toString().replace("-", "").substring(0, 8) + "_" + original; Path dir = Paths.get(System.getProperty("user.dir"), "uploads", "classroom"); Files.createDirectories(dir); file.transferTo(dir.resolve(stored).toFile());
        ClassroomMaterial row = new ClassroomMaterial().setClassroomId(classroomId).setFolderId(folderId == null ? 0L : folderId).setFileName(original).setFileType(extension(original)).setFilePath("/uploads/classroom/" + stored).setFileSize(file.getSize()).setCreatorId(access.uid()).setIsShared(shared == null ? 0 : shared).setDownloadCount(0).setStatus(1).setCreateTime(new Date()).setUpdateTime(new Date()); materialMapper.insert(row); return row;
    }
    public List<Map<String, Object>> materials(Long classroomId, Long folderId) throws Exception { viewable(classroomId); QueryWrapper<ClassroomMaterial> q = new QueryWrapper<ClassroomMaterial>().eq("classroom_id", classroomId).eq("status", 1).orderByDesc("create_time"); if (folderId != null && folderId != 0) q.eq("folder_id", folderId); return materialMapper.selectList(q).stream().map(this::materialView).collect(Collectors.toList()); }
    public void deleteMaterial(Long id) throws Exception { ClassroomMaterial row = material(id); access.requireManager(row.getClassroomId()); removeFile(row.getFilePath()); materialMapper.update(null, new UpdateWrapper<ClassroomMaterial>().eq("id", id).set("status", 0).set("update_time", new Date())); }
    public ClassroomMaterial copy(Map<String, Object> request) throws Exception { ClassroomMaterial source = material(number(request.get("materialId"))); ClassroomFolder target = folder(number(request.get("folderId"))); access.requireManager(target.getClassroomId()); ClassroomMaterial row = new ClassroomMaterial().setClassroomId(target.getClassroomId()).setFolderId(target.getId()).setFileName(source.getFileName()).setFileType(source.getFileType()).setFilePath(source.getFilePath()).setFileSize(source.getFileSize()).setCreatorId(access.uid()).setIsShared(0).setDownloadCount(0).setStatus(1).setCreateTime(new Date()).setUpdateTime(new Date()); materialMapper.insert(row); return row; }

    public List<Map<String, Object>> permissions(Long materialId) throws Exception { ClassroomMaterial material = material(materialId); access.requireManager(material.getClassroomId()); Map<String, ClassroomMaterialPermission> current = permissionMapper.selectList(new QueryWrapper<ClassroomMaterialPermission>().eq("material_id", materialId)).stream().collect(Collectors.toMap(ClassroomMaterialPermission::getStudentUid, p -> p)); List<Map<String, Object>> result = new ArrayList<>(); for (ClassroomStudent student : studentMapper.selectList(new QueryWrapper<ClassroomStudent>().eq("classroom_id", material.getClassroomId()).eq("status", 1))) { Map<String, Object> row = new LinkedHashMap<>(); row.put("uid", student.getUid()); row.put("realName", student.getRealName()); UserInfo user = userService.getById(student.getUid()); row.put("username", user == null ? "" : user.getUsername()); ClassroomMaterialPermission p = current.get(student.getUid()); row.put("canPreview", p != null && Objects.equals(p.getCanPreview(), 1)); row.put("canDownload", p != null && Objects.equals(p.getCanDownload(), 1)); result.add(row); } return result; }
    @Transactional(rollbackFor = Exception.class)
    public void setPermissions(Long materialId, List<Map<String, Object>> rows) throws Exception { ClassroomMaterial material = material(materialId); access.requireManager(material.getClassroomId()); permissionMapper.delete(new QueryWrapper<ClassroomMaterialPermission>().eq("material_id", materialId)); for (Map<String, Object> item : rows) { String uid = text(item.get("uid")); if (uid.isEmpty()) continue; permissionMapper.insert(new ClassroomMaterialPermission().setMaterialId(materialId).setStudentUid(uid).setCanPreview(bool(item.get("canPreview"))).setCanDownload(bool(item.get("canDownload"))).setCreateTime(new Date()).setUpdateTime(new Date())); } }
    public void setAllPermissions(Long materialId, boolean preview, boolean download) throws Exception { ClassroomMaterial material = material(materialId); access.requireManager(material.getClassroomId()); List<Map<String, Object>> rows = new ArrayList<>(); for (ClassroomStudent s : studentMapper.selectList(new QueryWrapper<ClassroomStudent>().eq("classroom_id", material.getClassroomId()).eq("status", 1))) { Map<String, Object> row = new HashMap<>(); row.put("uid", s.getUid()); row.put("canPreview", preview); row.put("canDownload", download); rows.add(row); } setPermissions(materialId, rows); }
    public boolean allowed(Long materialId, boolean download) throws Exception { ClassroomMaterial material = material(materialId); if (access.canManage(access.requireClassroom(material.getClassroomId()), access.uid()) || Objects.equals(material.getCreatorId(), access.uid())) return true; ClassroomMaterialPermission p = permissionMapper.selectOne(new QueryWrapper<ClassroomMaterialPermission>().eq("material_id", materialId).eq("student_uid", access.uid())); return p != null && Objects.equals(download ? p.getCanDownload() : p.getCanPreview(), 1); }
    public ClassroomMaterial material(Long id) throws StatusNotFoundException { ClassroomMaterial row = id == null ? null : materialMapper.selectOne(new QueryWrapper<ClassroomMaterial>().eq("id", id).eq("status", 1)); if (row == null) throw new StatusNotFoundException("资料不存在"); return row; }
    public File file(Long id, boolean download) throws Exception { if (!allowed(id, download)) throw new StatusForbiddenException(download ? "您没有下载权限" : "您没有预览权限"); ClassroomMaterial row = material(id); if (download) materialMapper.update(null, new UpdateWrapper<ClassroomMaterial>().eq("id", id).setSql("download_count = download_count + 1")); File file = new File(System.getProperty("user.dir"), row.getFilePath().replaceFirst("^/", "")); if (!file.isFile()) throw new StatusNotFoundException("文件不存在"); return file; }
    public Map<String, Object> previewToken(Long id) throws Exception { if (!allowed(id, false)) throw new StatusForbiddenException("您没有预览权限"); long expire = System.currentTimeMillis() / 1000 + 300; String value = id + "|" + access.uid() + "|" + expire; String token = base64(value + "|" + digest(value)); Map<String, Object> result = new LinkedHashMap<>(); result.put("token", token); result.put("previewUrl", "/api/classroom/material/preview/" + token); return result; }
    public File previewByToken(String token) throws Exception { String raw; try { raw = new String(Base64.getUrlDecoder().decode(token), "UTF-8"); } catch (Exception e) { throw new StatusForbiddenException("预览令牌无效"); } String[] parts = raw.split("\\|", -1); if (parts.length != 4 || Long.parseLong(parts[2]) < System.currentTimeMillis() / 1000 || !digest(parts[0] + "|" + parts[1] + "|" + parts[2]).equals(parts[3])) throw new StatusForbiddenException("预览令牌无效或已过期"); return material(Long.valueOf(parts[0])) == null ? null : new File(System.getProperty("user.dir"), material(Long.valueOf(parts[0])).getFilePath()); }

    public Map<String, Object> randomPick(Long classroomId) throws Exception { access.requireManager(classroomId); List<ClassroomStudent> students = studentMapper.selectList(new QueryWrapper<ClassroomStudent>().eq("classroom_id", classroomId).eq("status", 1)); if (students.isEmpty()) throw new StatusFailException("班级中没有学生"); ClassroomStudent student = students.get(new Random().nextInt(students.size())); ClassroomRandomPick row = new ClassroomRandomPick().setClassroomId(classroomId).setPickedUid(student.getUid()).setPickTime(new Date()); pickMapper.insert(row); return pickView(row, student); }
    public List<Map<String, Object>> pickHistory(Long classroomId, int page, int limit) throws Exception { viewable(classroomId); return pickMapper.selectList(new QueryWrapper<ClassroomRandomPick>().eq("classroom_id", classroomId).orderByDesc("pick_time").last("limit " + Math.min(100, Math.max(1, limit)) + " offset " + Math.max(0, page - 1) * Math.max(1, limit))).stream().map(row -> pickView(row, studentMapper.selectOne(new QueryWrapper<ClassroomStudent>().eq("classroom_id", classroomId).eq("uid", row.getPickedUid())))).collect(Collectors.toList()); }

    public Map<String, Object> sendMessage(Map<String, Object> request) throws Exception { Long classroomId = number(request.get("classroomId")); viewable(classroomId); String content = text(request.get("content")), image = text(request.get("imageUrl")); if (content.isEmpty() && image.isEmpty()) throw new StatusFailException("消息内容不能为空"); ClassroomMessage row = new ClassroomMessage().setClassroomId(classroomId).setSenderId(access.uid()).setContent(content).setImageUrl(image).setMsgType(image.isEmpty() ? "text" : "image").setCreateTime(new Date()); messageMapper.insert(row); return messageView(row); }
    public List<Map<String, Object>> messages(Long classroomId, int page, int limit) throws Exception { viewable(classroomId); return messageMapper.selectList(new QueryWrapper<ClassroomMessage>().eq("classroom_id", classroomId).orderByDesc("create_time").last("limit " + Math.min(100, Math.max(1, limit)) + " offset " + Math.max(0, page - 1) * Math.max(1, limit))).stream().sorted(Comparator.comparing(ClassroomMessage::getCreateTime)).map(this::messageView).collect(Collectors.toList()); }
    public void recallMessage(Long id) throws Exception { ClassroomMessage row = messageMapper.selectById(id); if (row == null) throw new StatusNotFoundException("消息不存在"); if (!Objects.equals(row.getSenderId(), access.uid()) && !access.isAdmin()) throw new StatusForbiddenException("无权撤回该消息"); messageMapper.deleteById(id); }
    public Map<String, Object> uploadImage(Long classroomId, MultipartFile file, String prefix) throws Exception { if (classroomId != null && classroomId != 0) viewable(classroomId); if (file == null || file.isEmpty()) throw new StatusFailException("请选择图片"); if (file.getSize() > 10 * 1024 * 1024L) throw new StatusFailException("图片大小不能超过10MB"); String ext = extension(file.getOriginalFilename()); if (!Arrays.asList("jpg", "jpeg", "png", "gif", "webp").contains(ext)) throw new StatusFailException("图片格式不支持"); String directory = prefix.startsWith("question_") ? "questions" : "images"; Path dir = Paths.get(System.getProperty("user.dir"), "uploads", "classroom", directory); Files.createDirectories(dir); String stored = prefix + System.currentTimeMillis() + "_" + UUID.randomUUID().toString().substring(0, 8) + "." + ext; file.transferTo(dir.resolve(stored).toFile()); Map<String, Object> result = new LinkedHashMap<>(); result.put("url", "/uploads/classroom/" + directory + "/" + stored); result.put("imageUrl", "/uploads/classroom/" + directory + "/" + stored); return result; }

    private Map<String, Object> materialView(ClassroomMaterial row) { Map<String, Object> value = new LinkedHashMap<>(); value.put("id", row.getId()); value.put("classroomId", row.getClassroomId()); value.put("folderId", row.getFolderId()); value.put("fileName", row.getFileName()); value.put("fileType", row.getFileType()); value.put("filePath", row.getFilePath()); value.put("fileSize", row.getFileSize()); value.put("creatorId", row.getCreatorId()); value.put("isShared", row.getIsShared()); value.put("downloadCount", row.getDownloadCount()); value.put("status", row.getStatus()); return value; }
    private Map<String, Object> pickView(ClassroomRandomPick row, ClassroomStudent student) {
        Map<String, Object> value = new LinkedHashMap<>();
        value.put("id", row.getId());
        value.put("classroomId", row.getClassroomId());
        value.put("pickedUid", row.getPickedUid());
        value.put("pickTime", row.getPickTime());
        if (student != null) {
            value.put("realName", student.getRealName());
            value.put("studentClass", student.getStudentClass());
            value.put("studentNo", student.getStudentNo());
            value.put("gender", student.getGender());
            value.put("studentInfo", studentView(student));
        }
        UserInfo user = userService.getById(row.getPickedUid());
        if (user != null) value.put("pickedUser", userView(user));
        return value;
    }

    private Map<String, Object> messageView(ClassroomMessage row) {
        Map<String, Object> value = new LinkedHashMap<>();
        value.put("id", row.getId());
        value.put("classroomId", row.getClassroomId());
        value.put("senderId", row.getSenderId());
        value.put("content", row.getContent());
        value.put("imageUrl", row.getImageUrl());
        value.put("msgType", row.getMsgType());
        value.put("createTime", row.getCreateTime());
        value.put("createdAt", row.getCreateTime());
        UserInfo user = userService.getById(row.getSenderId());
        if (user != null) {
            value.put("username", user.getUsername());
            value.put("sender", userView(user));
        }
        ClassroomStudent student = studentMapper.selectOne(
                new QueryWrapper<ClassroomStudent>()
                        .eq("classroom_id", row.getClassroomId())
                        .eq("uid", row.getSenderId())
                        .eq("status", 1));
        if (student != null) value.put("studentInfo", studentView(student));
        return value;
    }

    private ClassroomUserVO userView(UserInfo user) {
        return new ClassroomUserVO()
                .setUid(user.getUuid())
                .setUuid(user.getUuid())
                .setUsername(user.getUsername())
                .setNickname(user.getNickname())
                .setRealname(user.getRealname())
                .setAvatar(user.getAvatar());
    }

    private Map<String, Object> studentView(ClassroomStudent student) {
        Map<String, Object> value = new LinkedHashMap<>();
        value.put("id", student.getId());
        value.put("classroomId", student.getClassroomId());
        value.put("uid", student.getUid());
        value.put("realName", student.getRealName());
        value.put("gender", student.getGender());
        value.put("studentClass", student.getStudentClass());
        value.put("studentNo", student.getStudentNo());
        value.put("status", student.getStatus());
        return value;
    }
    private void viewable(Long classroomId) throws Exception { if (classroomId == null) throw new StatusFailException("班级编号不能为空"); if (!access.canManage(access.requireClassroom(classroomId), access.uid()) && !access.isStudent(classroomId, access.uid())) throw new StatusForbiddenException("无权访问该班级"); }
    private ClassroomFolder folder(Long id) throws StatusNotFoundException { ClassroomFolder row = id == null ? null : folderMapper.selectOne(new QueryWrapper<ClassroomFolder>().eq("id", id).eq("status", 1)); if (row == null) throw new StatusNotFoundException("文件夹不存在"); return row; }
    private List<Long> descendants(Long root) { List<Long> ids = new ArrayList<>(); ids.add(root); for (int i = 0; i < ids.size(); i++) ids.addAll(folderMapper.selectList(new QueryWrapper<ClassroomFolder>().eq("parent_id", ids.get(i)).eq("status", 1)).stream().map(ClassroomFolder::getId).collect(Collectors.toList())); return ids; }
    private void removeFile(String path) { if (path != null) try { Files.deleteIfExists(Paths.get(System.getProperty("user.dir"), path.replaceFirst("^/", ""))); } catch (IOException ignored) {} }
    private String extension(String name) { String value = name == null ? "" : name.toLowerCase(Locale.ROOT); int dot = value.lastIndexOf('.'); return dot < 0 ? "" : value.substring(dot + 1); }
    private int bool(Object value) { return value instanceof Boolean ? ((Boolean) value) ? 1 : 0 : "1".equals(String.valueOf(value)) ? 1 : 0; }
    private Long number(Object value) { if (value instanceof Number) return ((Number) value).longValue(); try { return value == null ? null : Long.parseLong(String.valueOf(value)); } catch (Exception e) { return null; } }
    private String text(Object value) { return value == null ? "" : String.valueOf(value); }
    private String base64(String value) { try { return Base64.getUrlEncoder().withoutPadding().encodeToString(value.getBytes("UTF-8")); } catch (Exception e) { throw new IllegalStateException(e); } }
    private String digest(String value) { try { byte[] bytes = MessageDigest.getInstance("SHA-256").digest(("histoj-preview-token-secret" + value).getBytes("UTF-8")); StringBuilder out = new StringBuilder(); for (byte b : bytes) out.append(String.format("%02x", b)); return out.toString(); } catch (Exception e) { throw new IllegalStateException(e); } }
}
