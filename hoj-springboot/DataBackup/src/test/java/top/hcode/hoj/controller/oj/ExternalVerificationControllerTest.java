package top.hcode.hoj.controller.oj;

import org.junit.jupiter.api.Test;
import org.springframework.mock.web.MockHttpServletRequest;
import org.springframework.mock.web.MockHttpServletResponse;

import static org.junit.jupiter.api.Assertions.*;

class ExternalVerificationControllerTest {
    @Test
    void untrustedClientCannotCreateNewRateLimitBucketsWithForwardedHeaders() {
        MockHttpServletRequest request = new MockHttpServletRequest();
        request.setRemoteAddr("203.0.113.1");
        request.addHeader("X-Real-IP", "198.51.100.1");
        request.addHeader("X-Forwarded-For", "198.51.100.2");
        assertEquals("203.0.113.1", ExternalVerificationController.clientIp(request));
        request.setRemoteAddr("127.0.0.1");
        assertEquals("198.51.100.1", ExternalVerificationController.clientIp(request));
        request.removeHeader("X-Real-IP");
        assertEquals("198.51.100.2", ExternalVerificationController.clientIp(request));
        request.removeHeader("X-Forwarded-For");
        assertEquals("127.0.0.1", ExternalVerificationController.clientIp(request));
    }

    @Test
    void nestedProxiesUseLastPublicPeerAndIgnoreSpoofedPrefix() {
        MockHttpServletRequest request = new MockHttpServletRequest();
        request.setRemoteAddr("172.18.0.5");
        request.addHeader("X-Real-IP", "172.18.0.1");
        request.addHeader("X-Forwarded-For", "198.51.100.99, 203.0.113.7, 172.18.0.1");
        assertEquals("203.0.113.7", ExternalVerificationController.clientIp(request));
        request.removeHeader("X-Forwarded-For");
        request.addHeader("X-Forwarded-For", "198.51.100.99, invalid-peer");
        assertEquals("172.18.0.5", ExternalVerificationController.clientIp(request));
    }

    @Test
    void malformedForwardedAddressFallsBackToPeer() {
        MockHttpServletRequest request = new MockHttpServletRequest();
        request.setRemoteAddr("127.0.0.1");
        request.addHeader("X-Real-IP", "attacker.example");
        assertEquals("127.0.0.1", ExternalVerificationController.clientIp(request));
    }

    @Test
    void anonymousStatementAndCodeResponsesCannotBeCachedOrForwardReferrers() {
        MockHttpServletResponse response = new MockHttpServletResponse();
        new ExternalVerificationController().privateResponse(response);
        assertEquals("no-store", response.getHeader("Cache-Control"));
        assertEquals("no-referrer", response.getHeader("Referrer-Policy"));
        assertEquals("noindex, nofollow", response.getHeader("X-Robots-Tag"));
    }
}
