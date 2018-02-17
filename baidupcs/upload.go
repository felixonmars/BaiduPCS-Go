package baidupcs

import (
	"fmt"
	"github.com/json-iterator/go"
	"net/http/cookiejar"
)

// RapidUpload 秒传文件
func (p *PCSApi) RapidUpload(targetPath, contentMD5, sliceMD5, crc32 string, length int64) (err error) {
	operation := "秒传文件"

	if targetPath == "/" {
		return fmt.Errorf("%s 遇到错误, 保存路径不能是根目录", operation)
	}

	p.setAPI("file", "rapidupload", map[string]string{
		"path":           targetPath,         // 上传文件的全路径名
		"content-length": fmt.Sprint(length), // 待秒传的文件长度
		"content-md5":    contentMD5,         // 待秒传的文件的MD5
		"slice-md5":      sliceMD5,           // 待秒传的文件的MD5
		"content-crc32":  crc32,              // 待秒传文件CRC32
		"ondup":          "overwrite",        // overwrite: 表示覆盖同名文件; newcopy: 表示生成文件副本并进行重命名，命名规则为“文件名_日期.后缀”
	})

	resp, err := p.client.Req("POST", p.url.String(), nil, nil)
	if err != nil {
		return
	}

	defer resp.Body.Close()

	errInfo := NewErrorInfo(operation)

	d := jsoniter.NewDecoder(resp.Body)
	err = d.Decode(errInfo)
	if err != nil {
		return fmt.Errorf("%s, json 数据解析失败, %s", operation, err)
	}

	switch errInfo.ErrCode {
	case 31079:
		// file md5 not found, you should use upload api to upload the whole file.
	}

	if errInfo.ErrCode != 0 {
		return errInfo
	}

	return nil
}

// Upload 上传单个文件
func (p *PCSApi) Upload(targetPath string, uploadFunc func(uploadURL string, jar *cookiejar.Jar) error) (err error) {
	if targetPath == "/" {
		return fmt.Errorf("上传文件 遇到错误, 保存路径不能是根目录")
	}

	p.setAPI("file", "upload", map[string]string{
		"path":  targetPath,
		"ondup": "overwrite",
	})

	return uploadFunc(p.url.String(), p.client.Jar.(*cookiejar.Jar))
}

// UploadTmpFile 分片上传—文件分片及上传
func (p *PCSApi) UploadTmpFile(targetPath string, uploadFunc func(uploadURL string, jar *cookiejar.Jar) error) (err error) {
	p.setAPI("file", "upload", map[string]string{
		"type": "tmpfile",
	})

	return uploadFunc(p.url.String(), p.client.Jar.(*cookiejar.Jar))
}

// UploadCreateSuperFile 分片上传—合并分片文件
func (p *PCSApi) UploadCreateSuperFile(targetPath string, blockList ...string) (err error) {
	operation := "分片上传—合并分片文件"

	if targetPath == "/" {
		return fmt.Errorf("%s 遇到错误, 保存路径不能是根目录", operation)
	}

	bl := struct {
		BlockList []string `json:"block_list"`
	}{
		BlockList: blockList,
	}

	data, err := jsoniter.Marshal(&bl)
	if err != nil {
		panic(operation + " 发生错误, " + err.Error())
	}

	p.setAPI("file", "createsuperfile", map[string]string{
		"path":  targetPath,
		"param": string(data),
		"ondup": "overwrite",
	})

	resp, err := p.client.Req("POST", p.url.String(), nil, nil)
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	errInfo := NewErrorInfo(operation)

	d := jsoniter.NewDecoder(resp.Body)
	err = d.Decode(errInfo)
	if err != nil {
		return fmt.Errorf("%s, json 数据解析失败, %s", operation, err)
	}

	if errInfo.ErrCode != 0 {
		return errInfo
	}

	return nil
}
