package com.grtsinry43.grtblog.service;

import com.baomidou.mybatisplus.extension.service.IService;
import com.grtsinry43.grtblog.dto.AddCategory;
import com.grtsinry43.grtblog.entity.Category;
import com.grtsinry43.grtblog.vo.CategoryVO;

import java.util.List;

/**
 * @author grtsinry43
 * @date 2024/11/12 16:11
 * @description 热爱可抵岁月漫长
 */
public interface ICategoryService extends IService<Category> {
    Boolean isCategoryExist(Long categoryId);

    CategoryVO addNewCategory(AddCategory addCategory);

    Long getOrCreateCategoryId(String name);

    String removeCategory(Long categoryId);

    List<CategoryVO> listAllCategories();

    Long getCategoryIdByShortUrl(String shortUrl);

    List<String> getAllCategoryShortLinks();

    Category getCategoryByShortUrl(String shortUrl);

    String getShortUrlById(Long id);
}
