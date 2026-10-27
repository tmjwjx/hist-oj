package top.hcode.hoj.filter;

import org.springframework.core.Ordered;
import org.springframework.core.annotation.Order;
import org.springframework.stereotype.Component;
import org.springframework.web.filter.OncePerRequestFilter;

import javax.servlet.FilterChain;
import javax.servlet.ReadListener;
import javax.servlet.ServletException;
import javax.servlet.ServletInputStream;
import javax.servlet.http.HttpServletRequest;
import javax.servlet.http.HttpServletRequestWrapper;
import javax.servlet.http.HttpServletResponse;
import java.io.BufferedReader;
import java.io.IOException;
import java.io.InputStreamReader;

@Component
@Order(Ordered.HIGHEST_PRECEDENCE)
public class ExternalVerificationRequestSizeFilter extends OncePerRequestFilter {
    private static final int MAX_BYTES = 512 * 1024;

    @Override
    protected boolean shouldNotFilter(HttpServletRequest request) {
        return !"POST".equalsIgnoreCase(request.getMethod())
                || !request.getServletPath().startsWith("/api/external-verification/");
    }

    @Override
    protected void doFilterInternal(HttpServletRequest request, HttpServletResponse response, FilterChain chain)
            throws ServletException, IOException {
        if (request.getContentLengthLong() > MAX_BYTES) {
            reject(response);
            return;
        }
        LimitedInputStream input = new LimitedInputStream(request.getInputStream());
        HttpServletRequestWrapper limited = new HttpServletRequestWrapper(request) {
            @Override public ServletInputStream getInputStream() { return input; }
            @Override public BufferedReader getReader() throws IOException {
                String encoding = getCharacterEncoding();
                return new BufferedReader(new InputStreamReader(input, encoding == null ? "UTF-8" : encoding));
            }
        };
        try {
            chain.doFilter(limited, response);
        } catch (IOException | ServletException exception) {
            if (!input.tooLarge) throw exception;
            reject(response);
        }
    }

    private void reject(HttpServletResponse response) throws IOException {
        if (response.isCommitted()) return;
        response.resetBuffer();
        response.setStatus(HttpServletResponse.SC_REQUEST_ENTITY_TOO_LARGE);
        response.setContentType("application/json;charset=UTF-8");
        response.setHeader("Cache-Control", "no-store");
        response.getWriter().write("{\"status\":413,\"data\":null,\"msg\":\"请求内容不能超过 512 KiB\"}");
    }

    private static class LimitedInputStream extends ServletInputStream {
        private final ServletInputStream input;
        private int consumed;
        private boolean tooLarge;

        private LimitedInputStream(ServletInputStream input) { this.input = input; }

        @Override public int read() throws IOException {
            count(0);
            int value = input.read();
            count(value == -1 ? 0 : 1);
            return value;
        }

        @Override public int read(byte[] bytes, int offset, int length) throws IOException {
            count(0);
            int read = input.read(bytes, offset, Math.min(length, MAX_BYTES - consumed + 1));
            count(Math.max(0, read));
            return read;
        }

        private void count(int size) throws IOException {
            consumed += size;
            if (consumed > MAX_BYTES) {
                tooLarge = true;
                throw new IOException("Request body exceeds 524288 bytes");
            }
        }

        @Override public boolean isFinished() { return input.isFinished(); }
        @Override public boolean isReady() { return input.isReady(); }
        @Override public void setReadListener(ReadListener listener) { input.setReadListener(listener); }
        @Override public void close() throws IOException { input.close(); }
    }
}
