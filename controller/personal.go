package controller

import (
	"errors"
	"fmt"
	"github.com/e421083458/golang_common/lib"
	"github.com/gin-gonic/gin"
	"wechatGin/common"
	"wechatGin/dao"
	"wechatGin/dto"
	"wechatGin/middleware"
	"wechatGin/public"
	"wechatGin/services"
)

type PersonController struct {
}

func PersonRegister(group *gin.RouterGroup) {
	personalController := &PersonController{}
	group.GET("/getBaseInfo", personalController.GetBaseInfo)               // 拿到基础信息
	group.GET("/getPersonSwapping", personalController.GetPersonalSwapping) // 正在进行轮播图
	group.GET("/noticeEmail", personalController.NoticeEmail)               // 提交反馈
	group.GET("/getUpToken", personalController.GetUpToken)                 // 拿到七牛云上传Token
	group.GET("/getCountData", personalController.GetCountData)             // 获取现在数量
}

func (person *PersonController) GetBaseInfo(c *gin.Context) {
	params := &dto.FormRegisteredInput{}
	if err := params.BindValidParam(c); err != nil {
		middleware.ResponseError(c, 4001, err) // 4001 参数不正确
		return
	}
	// todo 通过上下文查找
	nickName, err := GetInfoByContext("nickName", c)
	if err != nil {
		middleware.ResponseError(c, 4996, errors.New("token解析错误"))
		return
	}
	avatar, err := GetInfoByContext("avatar", c)
	if err != nil {
		middleware.ResponseError(c, 4997, errors.New("token解析错误"))
		return
	}
	identity, err := GetInfoByContext("identity", c)
	if err != nil {
		middleware.ResponseError(c, 4998, errors.New("token解析错误"))
		return
	}
	resInfo := dto.BaseInfoOutput{
		ImgUrl:   avatar,
		NickName: nickName,
		Identity: identity,
	}
	middleware.ResponseSuccess(c, resInfo)
}

func (person *PersonController) GetPersonalSwapping(c *gin.Context) {
	params := &dto.FormRegisteredInput{}
	if err := params.BindValidParam(c); err != nil {
		middleware.ResponseError(c, 4001, err) // 4001 参数不正确
		return
	}
	// todo 通过缓存找到这个人的uid
	uid, err := GetInfoByContext("uid", c)
	if err != nil {
		middleware.ResponseError(c, 4999, errors.New("token解析错误"))
		return
	}
	// todo 通过uid来查询
	info, err := services.GetStillDoing(uid)
	if err != nil {
		middleware.ResponseError(c, 4002, errors.New("服务器错误"))
		return
	}
	if len(info) > 4 { // 如果超过了4份
		info = info[:4] // 因为是个人中心没必要这么多
	}
	if len(info) == 0 {
		middleware.ResponseSuccess(c, "") // 没有就直接返回空
		return
	}
	// todo 拿到详细信息
	forms := make([]dao.Form, len(info))
	for k, _ := range forms {
		forms[k].RandomId = info[k]
	}
	cacheInfo, err := services.GetCacheInfo(uid, forms)
	if err != nil {
		middleware.ResponseError(c, 4003, errors.New("服务器错误"))
		return
	}
	middleware.ResponseSuccess(c, packageInfo(cacheInfo))
}

func (person *PersonController) NoticeEmail(c *gin.Context) {
	params := &dto.EmailNoticeInput{}
	if err := params.BindValidParam(c); err != nil {
		middleware.ResponseError(c, 4001, err)
		return
	}
	// todo 通过缓存找到这个人的uid
	uid, err := GetInfoByContext("uid", c)
	if err != nil {
		middleware.ResponseError(c, 4999, errors.New("token解析错误"))
		return
	}
	// todo 判断是否存在
	if common.BloomFilterService.Contains(uid) {
		middleware.ResponseError(c, 4004, errors.New("发送过于频繁"))
		return
	}
	// todo 加入Redis的消息队列
	resInfo := fmt.Sprintf("%v%v%v%v%v%v%v", uid, public.SplitSymbol, params.Email, public.SplitSymbol, params.Title, public.SplitSymbol, params.Title)
	err = pushRedisList(resInfo)
	if err != nil {
		middleware.ResponseError(c, 4005, errors.New("加入消息队列错误"))
		return
	}
	common.BloomFilterService.Add(uid)
	middleware.ResponseSuccess(c, "我们收到你的反馈啦")
}

func (person *PersonController) GetUpToken(c *gin.Context) {
	params := &dto.QiNiuCloudInput{}
	if err := params.BindValidParam(c); err != nil {
		middleware.ResponseError(c, 4001, err)
		return
	}
	// todo 通过缓存找到这个人的uid
	uid, err := GetInfoByContext("uid", c)
	if err != nil {
		middleware.ResponseError(c, 4999, errors.New("token解析错误"))
		return
	}
	// todo 验证文件名称
	strArray, flag := common.JudgeFileName(params.FileName)
	if !flag {
		middleware.ResponseError(c, 4006, errors.New("暂时不支持文件格式"))
		return
	}
	// todo 对文件名加上特征
	fileName := common.HandleFileName(strArray[0], uid)
	fileName = fileName + "." + strArray[1]
	token := common.GetQiNiuCloudUpToken(fileName) // 拼接回去
	middleware.ResponseSuccess(c, dto.QiNiuCloudTokenOutput{
		Token:    token,
		FileName: fileName,
	})
}

func (person *PersonController) GetCountData(c *gin.Context) {
	params := &dto.FormRegisteredInput{}
	if err := params.BindValidParam(c); err != nil {
		middleware.ResponseError(c, 4001, err) // 4001 参数不正确
		return
	}
	uid, err := GetInfoByContext("uid", c)
	if err != nil {
		middleware.ResponseError(c, 4999, errors.New("token解析错误"))
		return
	}
	tx, err := lib.GetGormPool("default")
	if err != nil {
		middleware.ResponseError(c, 10000, errors.New("服务器错误"))
		return
	}
	// todo 处理我创建的
	var form *dao.Form
	formCount, err := form.GetHowManyIDone(c, tx, uid)
	if err != nil {
		middleware.ResponseError(c, 4007, err)
		return
	}
	// todo 处理我拿到的
	var commit *dao.Commit
	commitCount, err := commit.GetHowManyIGet(tx, uid)
	if err != nil {
		middleware.ResponseError(c, 4008, err)
		return
	}
	middleware.ResponseSuccess(c, dto.CountDataOutput{
		IDoHowMany:  formCount,
		IGetHowMany: commitCount,
	})
}

func pushRedisList(stringInfo string) error {
	_, err := common.RedisConfDo("LPUSH", "email_timer", stringInfo)
	if err != nil {
		return err
	}
	return nil
}

func packageInfo(cacheInfo []dao.Form) []dto.DoingOutput {
	var res []dto.DoingOutput
	for k, _ := range cacheInfo {
		temp := dto.DoingOutput{}
		temp.Id = cacheInfo[k].RandomId
		temp.Title = cacheInfo[k].FormInfos.Title
		temp.Tip = cacheInfo[k].FormInfos.Tip
		temp.Time = cacheInfo[k].UpdatedAt.Format("2006-01-02 15:04")
		res = append(res, temp)
	}
	return res
}
