package top.hcode.hoj.pojo.dto;

import lombok.Data;

@Data
public class ProblemAiChatDTO {
    private Long pid;
    private String sessionId;
    private String message;
}
