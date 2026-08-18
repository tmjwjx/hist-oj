package top.hcode.hoj.manager.learning;

import com.baomidou.mybatisplus.core.conditions.query.QueryWrapper;
import org.springframework.stereotype.Component;
import org.springframework.transaction.annotation.Transactional;
import top.hcode.hoj.common.exception.LearningMapException;
import top.hcode.hoj.dao.learning.LearningMapEntityService;
import top.hcode.hoj.dao.learning.LearningMapPermissionEntityService;
import top.hcode.hoj.dao.user.UserInfoEntityService;
import top.hcode.hoj.pojo.entity.learning.LearningMap;
import top.hcode.hoj.pojo.entity.learning.LearningMapPermission;
import top.hcode.hoj.pojo.entity.user.UserInfo;
import top.hcode.hoj.pojo.vo.LearningMapAccessVO;
import top.hcode.hoj.pojo.vo.LearningMapPermissionItemVO;
import top.hcode.hoj.pojo.vo.LearningMapPermissionUserVO;

import javax.annotation.Resource;
import java.util.*;
import java.util.function.Function;
import java.util.stream.Collectors;

@Component
public class LearningMapPermissionManager {

    @Resource
    private LearningMapCatalogManager catalogManager;
    @Resource
    private LearningMapEntityService mapService;
    @Resource
    private LearningMapPermissionEntityService permissionService;
    @Resource
    private UserInfoEntityService userInfoService;

    public List<LearningMap> listAccessible(String uid) {
        List<LearningMap> maps = catalogManager.list(true);
        if (maps.isEmpty()) {
            return maps;
        }
        List<Long> mapIds = maps.stream().map(LearningMap::getId).collect(Collectors.toList());
        Map<Long, Boolean> overrides = permissionService.list(new QueryWrapper<LearningMapPermission>()
                        .eq("user_id", uid).in("map_id", mapIds)).stream()
                .collect(Collectors.toMap(LearningMapPermission::getMapId, LearningMapPermission::getEnabled));
        return maps.stream().filter(map -> allowed(map, overrides.get(map.getId()), overrides.containsKey(map.getId())))
                .collect(Collectors.toList());
    }

    public LearningMap ensureAccess(Long mapId, String uid) {
        LearningMap map = catalogManager.get(mapId);
        if (!LearningMapRules.PUBLISHED.equals(map.getStatus())) {
            throw LearningMapException.forbidden("航海图尚未发布");
        }
        LearningMapPermission permission = permissionService.getOne(new QueryWrapper<LearningMapPermission>()
                .eq("map_id", mapId).eq("user_id", uid));
        if (!allowed(map, permission == null ? null : permission.getEnabled(), permission != null)) {
            throw LearningMapException.forbidden("您没有权限访问该航海图");
        }
        return map;
    }

    public LearningMapAccessVO config(Long mapId) {
        LearningMap map = catalogManager.get(mapId);
        List<LearningMapPermission> permissions = permissionService.list(new QueryWrapper<LearningMapPermission>()
                .eq("map_id", mapId).orderByDesc("update_time"));
        Set<String> uids = permissions.stream().map(LearningMapPermission::getUserId).collect(Collectors.toSet());
        Map<String, UserInfo> users = uids.isEmpty() ? Collections.emptyMap()
                : userInfoService.listByIds(uids).stream().collect(Collectors.toMap(UserInfo::getUuid, Function.identity()));
        List<LearningMapPermissionItemVO> items = permissions.stream().map(permission -> {
            UserInfo user = users.get(permission.getUserId());
            return new LearningMapPermissionItemVO().setUserId(permission.getUserId())
                    .setUsername(user == null ? null : user.getUsername())
                    .setNickname(user == null ? null : user.getNickname())
                    .setRealname(user == null ? null : user.getRealname())
                    .setEnabled(permission.getEnabled()).setUpdatedAt(permission.getUpdateTime());
        }).collect(Collectors.toList());
        return new LearningMapAccessVO().setAccessMode(LearningMapRules.accessMode(map.getAccessMode()))
                .setPermissions(items);
    }

    public void setAccessMode(Long mapId, String mode) {
        LearningMap map = catalogManager.get(mapId).setAccessMode(LearningMapRules.accessMode(mode))
                .setUpdateTime(new Date());
        mapService.updateById(map);
    }

    @Transactional(rollbackFor = Exception.class)
    public void setPermission(Long mapId, String userId, boolean enabled) {
        catalogManager.get(mapId);
        String uid = LearningMapRules.text(userId);
        if (uid.isEmpty()) {
            throw new LearningMapException("用户ID不能为空");
        }
        LearningMapPermission permission = permissionService.getOne(new QueryWrapper<LearningMapPermission>()
                .eq("map_id", mapId).eq("user_id", uid));
        Date now = new Date();
        if (permission == null) {
            permission = new LearningMapPermission().setMapId(mapId).setUserId(uid).setEnabled(enabled)
                    .setCreateTime(now).setUpdateTime(now);
            permissionService.save(permission);
        } else {
            permission.setEnabled(enabled).setUpdateTime(now);
            permissionService.updateById(permission);
        }
    }

    @Transactional(rollbackFor = Exception.class)
    public void setPermissions(Long mapId, List<String> userIds, boolean enabled) {
        catalogManager.get(mapId);
        if (userIds == null) {
            return;
        }
        for (String userId : new LinkedHashSet<>(userIds)) {
            if (!LearningMapRules.text(userId).isEmpty()) {
                setPermission(mapId, userId, enabled);
            }
        }
    }

    public void deletePermission(Long mapId, String userId) {
        catalogManager.get(mapId);
        permissionService.remove(new QueryWrapper<LearningMapPermission>().eq("map_id", mapId)
                .eq("user_id", LearningMapRules.text(userId)));
    }

    public List<LearningMapPermissionUserVO> searchUsers(String keyword) {
        String value = LearningMapRules.text(keyword);
        if (value.isEmpty()) {
            return new ArrayList<>();
        }
        List<UserInfo> users = userInfoService.list(new QueryWrapper<UserInfo>()
                .and(wrapper -> wrapper.like("username", value).or().like("nickname", value).or().like("realname", value))
                .orderByAsc("uuid").last("LIMIT 20"));
        return users.stream().map(user -> new LearningMapPermissionUserVO().setUserId(user.getUuid())
                .setUsername(user.getUsername()).setNickname(user.getNickname()).setRealname(user.getRealname()))
                .collect(Collectors.toList());
    }

    private boolean allowed(LearningMap map, Boolean override, boolean hasOverride) {
        if (hasOverride) {
            return Boolean.TRUE.equals(override);
        }
        return LearningMapRules.ALL_OPEN.equals(LearningMapRules.accessMode(map.getAccessMode()));
    }
}
