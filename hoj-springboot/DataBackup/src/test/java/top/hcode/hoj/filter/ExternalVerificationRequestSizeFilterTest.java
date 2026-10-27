package top.hcode.hoj.filter;

import org.junit.jupiter.api.Test;
import org.springframework.mock.web.MockHttpServletRequest;
import org.springframework.mock.web.MockHttpServletResponse;
import org.springframework.util.StreamUtils;

import java.nio.charset.StandardCharsets;
import java.util.concurrent.atomic.AtomicInteger;

import static org.junit.jupiter.api.Assertions.*;

class ExternalVerificationRequestSizeFilterTest {
    private final ExternalVerificationRequestSizeFilter filter = new ExternalVerificationRequestSizeFilter();

    @Test
    void knownOversizeIsRejectedBeforeAnyBodyReadingOrControllerWork() throws Exception {
        MockHttpServletRequest request = new MockHttpServletRequest() {
            @Override public long getContentLengthLong() { return 512 * 1024 + 1; }
            @Override public javax.servlet.ServletInputStream getInputStream() {
                fail("Oversized body must not be opened");
                return null;
            }
        };
        request.setMethod("POST");
        request.setServletPath("/api/external-verification/token/submit");
        MockHttpServletResponse response = new MockHttpServletResponse();
        filter.doFilter(request, response, (req, res) -> fail("Oversized request reached controller"));
        assertEquals(413, response.getStatus());
        assertTrue(response.getContentAsString().contains("\"status\":413"));
        assertEquals("no-store", response.getHeader("Cache-Control"));
    }

    @Test
    void chunkedBodyCannotReadPastLimitWithoutContentLength() throws Exception {
        MockHttpServletRequest request = new MockHttpServletRequest() {
            @Override public long getContentLengthLong() { return -1; }
        };
        request.setMethod("POST");
        request.setServletPath("/api/external-verification/token/submit");
        request.addHeader("Transfer-Encoding", "chunked");
        request.setContent(new byte[512 * 1024 + 1]);
        MockHttpServletResponse response = new MockHttpServletResponse();
        AtomicInteger delivered = new AtomicInteger();
        filter.doFilter(request, response, (req, res) -> {
            byte[] buffer = new byte[8192];
            int count;
            while ((count = req.getInputStream().read(buffer)) != -1) delivered.addAndGet(count);
            fail("Chunked oversized body was accepted");
        });
        assertEquals(512 * 1024, delivered.get());
        assertEquals(413, response.getStatus());
    }

    @Test
    void normalUtf8JsonPassesThroughUnchanged() throws Exception {
        String body = "{\"nickname\":\"验题人\",\"code\":\"int main(){}\"}";
        MockHttpServletRequest request = new MockHttpServletRequest("POST", "/api/external-verification/token/submit");
        request.setServletPath("/api/external-verification/token/submit");
        request.setContent(body.getBytes(StandardCharsets.UTF_8));
        request.setContentType("application/json");
        MockHttpServletResponse response = new MockHttpServletResponse();
        filter.doFilter(request, response, (req, res) ->
                assertEquals(body, StreamUtils.copyToString(req.getInputStream(), StandardCharsets.UTF_8)));
        assertEquals(200, response.getStatus());
    }
}
