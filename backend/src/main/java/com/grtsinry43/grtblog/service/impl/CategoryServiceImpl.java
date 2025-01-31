package com.grtsinry43.grtblog.service.impl;

import com.baomidou.mybatisplus.core.conditions.query.QueryWrapper;
import com.baomidou.mybatisplus.extension.service.impl.ServiceImpl;
import com.grtsinry43.grtblog.dto.AddCategory;
import com.grtsinry43.grtblog.entity.Category;
import com.grtsinry43.grtblog.mapper.CategoryMapper;
import com.grtsinry43.grtblog.service.ICategoryService;
import com.grtsinry43.grtblog.util.ArticleParser;
import com.grtsinry43.grtblog.vo.CategoryVO;
import org.springframework.stereotype.Service;

import java.util.List;
import java.util.stream.Collectors;

/**
 * @author grtsinry43
 * @date 2024/11/12 16:12
 * @description 热爱可抵岁月漫长
 */
@Service
public class CategoryServiceImpl extends ServiceImpl<CategoryMapper, Category> implements ICategoryService {
    @Override
    public Boolean isCategoryExist(Long categoryId) {
        return getById(categoryId) != null;
    }

    @Override
    public CategoryVO addNewCategory(AddCategory addCategory) {
        Category category = new Category();
        category.setName(addCategory.getName());
        category.setShortUrl(addCategory.getShortUrl());
        save(category);
        CategoryVO categoryVO = new CategoryVO();
        categoryVO.setId(category.getId().toString());
        categoryVO.setName(category.getName());
        return categoryVO;
    }

    @Override
    public Long getOrCreateCategoryId(String name) {
        Category category = getOne(new QueryWrapper<Category>().eq("name", name));
        if (category == null) {
            category = new Category();
            category.setName(name);
            category.setShortUrl(ArticleParser.generateShortUrl(name));
            category.setArticle(true);
            save(category);
        }
        return category.getId();
    }

    @Override
    public String removeCategory(Long categoryId) {
        removeById(categoryId);
        return "success";
    }

    @Override
    public List<CategoryVO> listAllCategories() {
        List<Category> categories = list();
        return categories.stream().map(category -> {
            CategoryVO categoryVO = new CategoryVO();
            categoryVO.setId(category.getId().toString());
            categoryVO.setName(category.getName());
            categoryVO.setShortUrl(category.getShortUrl());
            categoryVO.setIsArticle(category.isArticle());
            return categoryVO;
        }).collect(Collectors.toList());
    }

    @Override
    public Long getCategoryIdByShortUrl(String shortUrl) {
        Category category = getOne(new QueryWrapper<Category>().eq("short_url", shortUrl));
        return category != null ? category.getId() : null;
    }

    @Override
    public List<String> getAllCategoryShortLinks() {
        List<Category> categories = list();
        return categories.stream().map(Category::getShortUrl).collect(Collectors.toList());
    }

    @Override
    public Category getCategoryByShortUrl(String shortUrl) {
        return getOne(new QueryWrapper<Category>().eq("short_url", shortUrl));
    }

    @Override
    public String getShortUrlById(Long id) {
        Category category = getById(id);
        return category != null ? category.getShortUrl() : null;
    }
}
