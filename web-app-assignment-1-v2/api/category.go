package api

import (
	"a21hc3NpZ25tZW50/model"
	"a21hc3NpZ25tZW50/service"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type CategoryAPI interface {
	AddCategory(c *gin.Context)
	UpdateCategory(c *gin.Context)
	DeleteCategory(c *gin.Context)
	GetCategoryByID(c *gin.Context)
	GetCategoryList(c *gin.Context)
}

type categoryAPI struct {
	categoryService service.CategoryService
}

func NewCategoryAPI(categoryRepo service.CategoryService) *categoryAPI {
	return &categoryAPI{categoryRepo}
}

func (ct *categoryAPI) AddCategory(c *gin.Context) {
	var newCategory model.Category
	if err := c.ShouldBindJSON(&newCategory); err != nil {
		c.JSON(http.StatusBadRequest, model.ErrorResponse{Error: err.Error()})
		return
	}

	err := ct.categoryService.Store(&newCategory)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, model.SuccessResponse{Message: "add category success"})
}

func (ct *categoryAPI) UpdateCategory(c *gin.Context) {
    var updatedCategory model.Category
    if err := c.ShouldBindJSON(&updatedCategory); err != nil {
        c.JSON(http.StatusBadRequest, model.ErrorResponse{Error: err.Error()})
        return
    }

    categoryID, err := strconv.Atoi(c.Param("id"))
    if err != nil {
        c.JSON(http.StatusBadRequest, model.ErrorResponse{Error: "Invalid category ID"})
        return
    }

    _, err = ct.categoryService.GetByID(categoryID)
    if err != nil {
        if err.Error() == "record not found" {
            c.JSON(http.StatusNotFound, model.ErrorResponse{Error: "Category not found"})
            return
        }
        c.JSON(http.StatusInternalServerError, model.ErrorResponse{Error: err.Error()})
        return
    }

    updatedCategory.ID = categoryID
    err = ct.categoryService.Update(updatedCategory.ID, updatedCategory)
    if err != nil {
        c.JSON(http.StatusInternalServerError, model.ErrorResponse{Error: err.Error()})
        return
    }

    c.JSON(http.StatusOK, model.SuccessResponse{Message: "update category success"})
}

func (ct *categoryAPI) DeleteCategory(c *gin.Context) {
    categoryID, err := strconv.Atoi(c.Param("id"))
    if err != nil {
        c.JSON(http.StatusBadRequest, model.ErrorResponse{Error: "Invalid category ID"})
        return
    }

    err = ct.categoryService.Delete(categoryID)
    if err != nil {
        c.JSON(http.StatusInternalServerError, model.ErrorResponse{Error: err.Error()})
        return
    }

    c.JSON(http.StatusOK, model.SuccessResponse{Message: "delete category success"})
}


func (ct *categoryAPI) GetCategoryByID(c *gin.Context) {
	categoryID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, model.ErrorResponse{Error: "Invalid category ID"})
		return
	}

	category, err := ct.categoryService.GetByID(categoryID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, category)
}

func (ct *categoryAPI) GetCategoryList(c *gin.Context) {
    categories, err := ct.categoryService.GetList()
    if err != nil {
        c.JSON(http.StatusInternalServerError, model.ErrorResponse{Error: err.Error()})
        return
    }

    c.JSON(http.StatusOK, categories)
}

