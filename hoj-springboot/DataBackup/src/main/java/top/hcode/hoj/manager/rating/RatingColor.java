package top.hcode.hoj.manager.rating;

import java.util.LinkedHashMap;
import java.util.Map;

public final class RatingColor {
    private RatingColor() { }

    public static Map<String, Object> of(int rating) {
        String color = "#000000", name = "Unrated", nameZh = "未评级";
        if (rating > 0) {
            String[][] levels = {
                    {"1", "1199", "#808080", "Newbie", "新手"},
                    {"1200", "1399", "#008000", "Pupil", "学徒"},
                    {"1400", "1599", "#03A89E", "Specialist", "专家"},
                    {"1600", "1899", "#0000FF", "Expert", "专家"},
                    {"1900", "2099", "#AA00AA", "Candidate Master", "候选大师"},
                    {"2100", "2299", "#FF8C00", "Master", "大师"},
                    {"2300", "2399", "#FF8C00", "International Master", "国际大师"},
                    {"2400", "2599", "#FF0000", "Grandmaster", "特级大师"},
                    {"2600", "2999", "#FF0000", "International Grandmaster", "国际特级大师"},
                    {"3000", "3999", "#FF0000", "Legendary Grandmaster", "传奇特级大师"},
                    {"4000", "999999", "#FF0000", "Tourist", "Tourist"}
            };
            for (String[] level : levels) if (rating >= Integer.parseInt(level[0]) && rating <= Integer.parseInt(level[1])) {
                color = level[2]; name = level[3]; nameZh = level[4]; break;
            }
        }
        Map<String, Object> result = new LinkedHashMap<>();
        result.put("rating", rating); result.put("color", color); result.put("name", name);
        result.put("nameZh", nameZh); result.put("level", name); return result;
    }
}
