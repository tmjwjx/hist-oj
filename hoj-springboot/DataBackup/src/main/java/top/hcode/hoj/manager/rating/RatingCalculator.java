package top.hcode.hoj.manager.rating;

import java.util.*;

/** Codeforces 风格的站内 Rating 计算，保持 Go 扩展层的公式和修正顺序。 */
public final class RatingCalculator {
    private static final int SEARCH_LOW = -10000;
    private static final int SEARCH_HIGH = 10000;

    private RatingCalculator() { }

    public static Map<String, Integer> changes(List<Participant> source) {
        if (source == null || source.isEmpty()) return Collections.emptyMap();
        List<Participant> users = new ArrayList<>(source);
        users.sort(Comparator.comparingInt(Participant::getRank).thenComparing(Participant::getUid));
        int[] ratings = users.stream().mapToInt(Participant::getRating).toArray();
        for (int i = 0; i < users.size(); i++) {
            double seed = seed(users.get(i).getRating(), ratings, i);
            double midRank = Math.sqrt(users.get(i).getRank() * seed);
            users.get(i).setDelta((needRating(midRank, ratings, i) - users.get(i).getRating()) / 2);
        }
        correctionOne(users);
        correctionTwo(users);
        Map<String, Integer> result = new HashMap<>();
        for (Participant user : users) result.put(user.getUid(), user.getDelta());
        return result;
    }

    private static double seed(int rating, int[] all, int skip) {
        double value = 1D;
        for (int i = 0; i < all.length; i++) if (i != skip) value += winProbability(rating, all[i]);
        return value;
    }

    private static double winProbability(int rating, int other) {
        return 1D / (1D + Math.pow(10D, (rating - other) / 400D));
    }

    private static int needRating(double target, int[] all, int skip) {
        int left = SEARCH_LOW, right = SEARCH_HIGH;
        while (right - left > 1) {
            int mid = (left + right) / 2;
            if (seed(mid, all, skip) < target) right = mid; else left = mid;
        }
        return left;
    }

    private static void correctionOne(List<Participant> users) {
        int sum = users.stream().mapToInt(Participant::getDelta).sum();
        int increment = (int) Math.floor(-sum / (double) users.size() - 1D);
        for (Participant user : users) user.setDelta(user.getDelta() + increment);
    }

    private static void correctionTwo(List<Participant> users) {
        List<Participant> ordered = new ArrayList<>(users);
        ordered.sort(Comparator.comparingInt(Participant::getRating).reversed()
                .thenComparingInt(Participant::getRank));
        int topCount = Math.max(1, Math.min(users.size(), (int) (4D * Math.sqrt(users.size()))));
        int topSum = ordered.subList(0, topCount).stream().mapToInt(Participant::getDelta).sum();
        int increment = (int) Math.floor(-topSum / (double) topCount);
        increment = Math.max(-10, Math.min(0, increment));
        for (Participant user : users) user.setDelta(user.getDelta() + increment);
    }

    public static int newRating(int oldRating, int change) { return Math.max(0, oldRating + change); }

    @lombok.Data
    public static class Participant {
        private String uid;
        private int rank;
        private int rating;
        private int delta;
        public Participant(String uid, int rank, int rating) {
            this.uid = uid; this.rank = rank; this.rating = rating;
        }
    }
}
