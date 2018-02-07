package baidupcs

import (
	"encoding/json"
	"fmt"
	"github.com/bitly/go-simplejson"
	"net/http/cookiejar"
)

// RapidUpload 秒传文件
func (p *PCSApi) RapidUpload(targetPath, contentMD5, sliceMD5, crc32 string, length int64) (err error) {
	if targetPath == "/" {
		return fmt.Errorf("秒传文件 遇到错误, 保存路径不能是根目录\n")
	}

	p.setApi("file", "rapidupload", map[string]string{
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

	json, err := simplejson.NewFromReader(resp.Body)
	if err != nil {
		return
	}

	code, msg := CheckErr(json)
	switch code {
	case 31079:
		// file md5 not found, you should use upload api to upload the whole file.
	}

	if msg != "" {
		return fmt.Errorf("秒传文件 遇到错误, 错误代码: %d, 消息: %s", code, msg)
	}

	return nil
}

// Upload 上传单个文件
func (p *PCSApi) Upload(targetPath string, uploadFunc func(uploadURL string, jar *cookiejar.Jar) error) (err error) {
	if targetPath == "/" {
		return fmt.Errorf("上传文件 遇到错误, 保存路径不能是根目录\n")
	}

	p.setApi("file", "upload", map[string]string{
		"path":  targetPath,
		"ondup": "overwrite",
	})

	return uploadFunc(p.url.String(), p.client.Jar.(*cookiejar.Jar))
}

// UploadTmpFile 分片上传—文件分片及上传
func (p *PCSApi) UploadTmpFile(targetPath string, uploadFunc func(uploadURL string, jar *cookiejar.Jar) error) (err error) {
	p.setApi("file", "upload", map[string]string{
		"path": targetPath,
		"type": "tmpfile",
	})

	return uploadFunc(p.url.String(), p.client.Jar.(*cookiejar.Jar))
}

// UploadCreateSuperFile 分片上传—合并分片文件
func (p *PCSApi) UploadCreateSuperFile(targetPath string, blockList ...string) (err error) {
	bl := struct {
		BlockList []string `json:"block_list"`
	}{
		BlockList: blockList,
	}

	data, _ := json.Marshal(&bl)

	p.setApi("file", "createsuperfile", map[string]string{
		"path":  targetPath,
		"param": string(data),
		"ondup": "overwrite",
	})

	resp, err := p.client.Req("POST", p.url.String(), nil, nil)
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	json, err := simplejson.NewFromReader(resp.Body)
	if err != nil {
		return
	}

	code, msg := CheckErr(json)
	if msg != "" {
		return fmt.Errorf("合并分片文件 遇到错误, 错误代码: %d, 消息: %s", code, msg)
	}

	return nil
}
