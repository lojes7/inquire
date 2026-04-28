package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/lojes7/inquire/internal/model"
	"github.com/lojes7/inquire/internal/service"
	"github.com/lojes7/inquire/pkg/response"
	"github.com/lojes7/inquire/pkg/secure"
)

// SemanticSearchFiles 语义搜索工作区文件
// @Summary      语义搜索文件
// @Description  根据自然语言查询，在工作区文件中执行语义搜索
// @Tags         file
// @Accept       json
// @Produce      json
// @Param        Authorization header string true "Bearer Token"
// @Param        req  body      model.FileSemanticSearchReq  true  "语义搜索请求体"
// @Success      200  {object}  response.Response   "搜索结果"
// @Failure      400  {object}  response.Response   "json解析出错"
// @Failure      500  {object}  response.Response   "服务器错误"
// @Router       /auth/files/search [post]
func SemanticSearchFiles(c *gin.Context) {
	userID := c.GetUint64("id")

	var req model.FileSemanticSearchReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, 400, "json 解析出错")
		return
	}

	files, err := service.SearchFiles(userID, req.Query, req.Limit, req.Threshold)
	if err != nil {
		if myErr := secure.Unwrap(err); myErr != nil {
			response.Fail(c, myErr.Code, myErr.Message)
		} else {
			response.Fail(c, 500, "服务器错误")
		}
		return
	}

	resp := model.SemanticSearchFilesResp{Files: files}
	response.Success(c, 200, "success", resp)
}
