package oss

import (
	"fmt"
	openapi "github.com/alibabacloud-go/darabonba-openapi/v2/client"
	sts20150401 "github.com/alibabacloud-go/sts-20150401/v2/client"
	util "github.com/alibabacloud-go/tea-utils/v2/service"
	"github.com/alibabacloud-go/tea/tea"
	"github.com/aliyun/aliyun-oss-go-sdk/oss"
	"github.com/pkg/errors"
	"net/http"
	"onpaper-api-go/logger"
	"onpaper-api-go/settings"
	"strings"
)

func CreateClient(accessKeyId *string, accessKeySecret *string) (_result *sts20150401.Client, _err error) {
	config := &openapi.Config{
		// 必填，您的 AccessKey ID
		AccessKeyId: accessKeyId,
		// 必填，您的 AccessKey Secret
		AccessKeySecret: accessKeySecret,
	}
	// 访问的域名
	config.Endpoint = tea.String("sts.cn-shenzhen.aliyuncs.com")
	_result = &sts20150401.Client{}
	_result, _err = sts20150401.NewClient(config)
	return _result, _err
}

func CreateOssClient() (client *oss.Client, err error) {
	// 创建OSSClient实例。
	client, err = oss.New(settings.Conf.Endpoint, settings.Conf.OssMaxSecretId, settings.Conf.OssMaxSecretKey)
	if err != nil {
		return
	}
	return
}

func CreatSTS(userId, stsType string) (sts *sts20150401.AssumeRoleResponse, err error) {
	client, err := CreateClient(tea.String(settings.Conf.OssStsSecretId), tea.String(settings.Conf.OssStsSecretKey))
	if err != nil {
		return
	}
	bucket := settings.Conf.TempBucket
	if stsType == "messages" {
		bucket = settings.Conf.PreviewBucket
	}
	appId := settings.Conf.AppId
	assumeRoleRequest := &sts20150401.AssumeRoleRequest{
		RoleArn: tea.String(settings.Conf.StsRoleArn),
		Policy: tea.String(fmt.Sprintf(`{
    		"Version": "1",
			"Statement": [
				{
					"Effect": "Allow",
					"Action": "oss:PutObject",
					"Resource":  [
        				"acs:oss:*:%s:%s/%s/%s/*"
					]
				}
			]}`, appId, bucket, stsType, userId)),
		RoleSessionName: tea.String("browser"), // 自定义角色会话名称。
		DurationSeconds: tea.Int64(900),        // 过期时间秒数 最小900秒
	}

	runtime := &util.RuntimeOptions{}
	sts, err = client.AssumeRoleWithOptions(assumeRoleRequest, runtime)
	if err != nil {
		return
	}
	return
}

// MoveTempToOriginal 移动临时桶的文件到原始桶
func MoveTempToOriginal(dest string) (err error) {
	// 创建OSSClient实例。
	client, err := CreateOssClient()
	if err != nil {
		return
	}

	// 填写源Bucket名称
	srcBucketName := settings.Conf.TempBucket
	// 填写拷贝前文件的完整路径
	srcObjectName := dest
	// 填写拷贝后文件的完整路径
	dstObjectName := dest
	// 填写目标Bucket名称
	destBucketName := settings.Conf.OriginalBucket

	bucket, err := client.Bucket(destBucketName)
	if err != nil {
		return
	}

	//移动临时桶的文件到原始桶
	_, err = bucket.CopyObjectFrom(srcBucketName, srcObjectName, dstObjectName)

	return
}

// MoveTempToPreView 移动临时桶的文件到阅览桶
func MoveTempToPreView(dest string) (err error) {
	// 创建OSSClient实例。
	client, err := CreateOssClient()
	if err != nil {
		return
	}

	// 填写源Bucket名称
	srcBucketName := settings.Conf.TempBucket
	// 填写拷贝前文件的完整路径
	srcObjectName := dest
	// 填写拷贝后文件的完整路径
	dstObjectName := dest
	// 填写目标Bucket名称
	destBucketName := settings.Conf.PreviewBucket

	bucket, err := client.Bucket(destBucketName)
	if err != nil {
		return
	}

	//移动临时桶的文件到原始桶
	_, err = bucket.CopyObjectFrom(srcBucketName, srcObjectName, dstObjectName)

	return
}

// SelectOssFileInfo 查找文件信息
func SelectOssFileInfo(bucketName string, fileName string) (mate http.Header, err error) {
	client, err := CreateOssClient()
	if err != nil {
		return
	}
	bucket, err := client.Bucket(bucketName)
	if err != nil {
		return
	}

	// 获取文件元信息。
	mate, err = bucket.GetObjectDetailedMeta(fileName)
	if err != nil {
		return
	}

	return
}

// BatchDeleteOssObject 批量删除文件
func BatchDeleteOssObject(bucketName string, dir, exclude string) (err error) {
	// 前缀prefix的值为空字符串或者NULL，将会删除整个Bucket内的所有文件，请谨慎使用。
	if dir == "" {
		return errors.New("前缀prefix的值为空字符串或者NULL，将会删除整个Bucket内的所有文件")
	}

	client, err := CreateOssClient()
	if err != nil {
		return
	}
	bucket, err := client.Bucket(bucketName)
	if err != nil {
		return
	}

	// 列举所有包含指定前缀的文件并删除。
	marker := oss.Marker("")
	// 如果您仅需要删除src目录及目录下的所有文件，则prefix设置为src/。

	prefix := oss.Prefix(dir)
	for {
		// 列举目录下文件
		lor, _err := bucket.ListObjects(marker, prefix)
		if _err != nil {
			err = _err
			return
		}

		var objects []string
		for _, object := range lor.Objects {
			// 剔除的不删除的
			if find := strings.Contains(object.Key, exclude); find && exclude != "" {
				continue
			}
			objects = append(objects, object.Key)
		}
		// 将oss.DeleteObjectsQuiet设置为true，表示不返回删除结果。
		delRes, _err := bucket.DeleteObjects(objects, oss.DeleteObjectsQuiet(true))
		if _err != nil {
			err = _err
			return
		}

		if len(delRes.DeletedObjects) > 0 {
			_err = errors.New("these objects deleted failure")
			logger.ErrZapLog(_err, delRes.DeletedObjects)
		}

		prefix = oss.Prefix(lor.Prefix)
		marker = oss.Marker(lor.NextMarker)
		if !lor.IsTruncated {
			break
		}
	}

	return
}
