package top.hcode.hoj.controller.classroom;

import lombok.extern.slf4j.Slf4j;
import top.hcode.hoj.common.exception.StatusFailException;
import top.hcode.hoj.common.exception.StatusForbiddenException;
import top.hcode.hoj.common.exception.StatusNotFoundException;
import top.hcode.hoj.common.result.CommonResult;
import top.hcode.hoj.common.result.ResultStatus;

@Slf4j
abstract class ClassroomControllerSupport {
    protected <T> CommonResult<T> run(Action<T> action) {
        try {
            return CommonResult.successResponse(action.execute());
        } catch (StatusNotFoundException e) {
            return CommonResult.errorResponse(e.getMessage(), ResultStatus.NOT_FOUND);
        } catch (StatusForbiddenException e) {
            return CommonResult.errorResponse(e.getMessage(), ResultStatus.FORBIDDEN);
        } catch (StatusFailException e) {
            return CommonResult.errorResponse(e.getMessage(), ResultStatus.FAIL);
        } catch (Exception e) {
            log.error("classroom request failed", e);
            return CommonResult.errorResponse(e.getMessage() == null ? "课堂服务暂时不可用" : e.getMessage(),
                    ResultStatus.SYSTEM_ERROR);
        }
    }

    protected CommonResult<Void> runVoid(VoidAction action) {
        return run(() -> {
            action.execute();
            return null;
        });
    }

    @FunctionalInterface
    protected interface Action<T> {
        T execute() throws Exception;
    }

    @FunctionalInterface
    protected interface VoidAction {
        void execute() throws Exception;
    }
}
