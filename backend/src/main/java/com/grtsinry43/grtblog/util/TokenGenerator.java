package com.grtsinry43.grtblog.util;

import java.security.MessageDigest;
import java.security.NoSuchAlgorithmException;
import java.security.SecureRandom;

/**
 * @author grtsinry43
 * @date 2025/1/26 20:41
 * @description 热爱可抵岁月漫长
 */

public class TokenGenerator {

    public static class TokenPair {
        private final String token;
        private final String hashedToken;

        public TokenPair(String token, String hashedToken) {
            this.token = token;
            this.hashedToken = hashedToken;
        }

        public String getToken() {
            return token;
        }

        public String getHashedToken() {
            return hashedToken;
        }
    }

    public static TokenPair generateToken() {
        try {
            // Generate a random 32-byte token
            SecureRandom secureRandom = new SecureRandom();
            byte[] randomBytes = new byte[32];
            secureRandom.nextBytes(randomBytes);
            String token = bytesToHex(randomBytes);

            // Hash the token using SHA-256
            String hashedToken = calculateHashedToken(randomBytes);

            return new TokenPair(token, hashedToken);
        } catch (Exception e) {
            throw new RuntimeException("Error generating token", e);
        }
    }

    public static String calculateHashedToken(String token) {
        try {
            // Convert the token to bytes
            byte[] tokenBytes = hexStringToBytes(token);
            // Hash the provided token using SHA-256
            return calculateHashedToken(tokenBytes);
        } catch (Exception e) {
            throw new RuntimeException("Error calculating hashed token", e);
        }
    }

    private static String calculateHashedToken(byte[] tokenBytes) throws NoSuchAlgorithmException {
        MessageDigest digest = MessageDigest.getInstance("SHA-256");
        byte[] hashedBytes = digest.digest(tokenBytes);
        return bytesToHex(hashedBytes);
    }

    private static String bytesToHex(byte[] bytes) {
        StringBuilder hexString = new StringBuilder();
        for (byte b : bytes) {
            String hex = Integer.toHexString(0xff & b);
            if (hex.length() == 1) {
                hexString.append('0');
            }
            hexString.append(hex);
        }
        return hexString.toString();
    }

    private static byte[] hexStringToBytes(String hexString) {
        int len = hexString.length();
        byte[] data = new byte[len / 2];
        for (int i = 0; i < len; i += 2) {
            data[i / 2] = (byte) ((Character.digit(hexString.charAt(i), 16) << 4)
                    + Character.digit(hexString.charAt(i+1), 16));
        }
        return data;
    }
}