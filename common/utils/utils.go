package utils

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"errors"
	"github.com/foxsuagr-sanse/go-gobang_game/common/auth"
	"github.com/foxsuagr-sanse/go-gobang_game/common/config"
	"github.com/gin-gonic/gin"
	"github.com/ymzuiku/hit"
	"os"
	"strings"
)

func GinMatchToken(c * gin.Context) (*auth.MyClaims,bool) {
	tokenHeader :=c.Request.Header.Get("Authorization")
	tokenInfo := strings.SplitN(tokenHeader, " ", 2)
	var jwt auth.JwtAPI = &auth.JWT{}
	jwt.Init()
	if claims,bl := jwt.MatchToken(tokenInfo[1]);bl {
		return claims,bl
	} else {
		return nil, false
	}
}

func OpenConfig() *config.AutoGenerated {
	var con config.ConFig = &config.Config{}
	conf := con.InitConfig()
	return conf.ConfData
}

func MatchPubKeyAndPriKey() bool {
	pathHead, _ := os.Getwd()
	pubkeyfile, err := os.Open(pathHead + "/cache/rsa/public.pem")
	prikeyfile, err2 := os.Open(pathHead + "/cache/rsa/private.pem")
	// 检查文件是否存在
	if err != nil || err2 != nil {
		return false
	}
	defer pubkeyfile.Close()
	defer prikeyfile.Close()
	// 检查文件内容
	pubkeyfileInfo,_ := pubkeyfile.Stat()
	prikeyfileInfo, _ := prikeyfile.Stat()
	if pubkeyfileInfo.Size() != 280 || prikeyfileInfo.Size() < 887 {
		return false
	}
	return true
}

func NewRsaPublicKey() (string,error) {
	// 生成并返回一个可用的RSA公钥
	// 检测文件中有无生成好的公私钥
	if !MatchPubKeyAndPriKey() {
		// 重新创建文件中的内容
		pathHead, _ := os.Getwd()
		pubFile,_ := os.Create(pathHead + "/cache/rsa/public.pem")
		priFile,_ := os.Create(pathHead + "/cache/rsa/private.pem")
		defer pubFile.Close()
		defer priFile.Close()
		// 创建私钥
		private,_ := rsa.GenerateKey(rand.Reader,1024)
		// 获得公钥
		public := private.PublicKey
		// 使用x509标准转换为pem格式
		derText := x509.MarshalPKCS1PrivateKey(private)
		// 创建私钥结构体
		block := pem.Block{
			Type: "RSA PRIVATE KEY",
			Bytes: derText,
		}
		// 写入文件
		err1 := pem.Encode(priFile, &block)

		// 将公钥转换为pem格式
		derpText, _ := x509.MarshalPKIXPublicKey(&public)
		// 创建公钥结构体
		block = pem.Block{
			Type: "PUBLIC KEY",
			Bytes: derpText,
		}
		// 写入文件
		err2 := pem.Encode(pubFile,&block)
		// 返回公钥
		// 将公钥写入到内存
		buf := pem.EncodeToMemory(&block)
		bol := hit.If(err1 == nil && err2 == nil,true,false)
		if bol.(bool) {
			return string(buf),nil
		} else {
			return "",errors.New("创建密钥失败!")
		}
	}
	// 文件存在，内容正确执行
	pathHead, _ := os.Getwd()
	public,_ := os.Open(pathHead + "/cache/rsa/public.pem")
	defer public.Close()
	info,_ := public.Stat()
	buf := make([]byte,info.Size())
	_, err := public.Read(buf)
	return string(buf),err
}

func DecodeMessage(cipherText string) (string,error) {
	// 使用标准base64编码解码字符串
	src, err := base64.StdEncoding.DecodeString(cipherText)
	if err != nil {
		return "",errors.New("base64解码失败")
	}
	// 使用私钥解密密文
	pathHead, _ := os.Getwd()
	private,err := os.Open(pathHead + "/cache/rsa/private.pem")
	if err != nil {
		return "",errors.New("未找到私钥文件")
	}
	defer private.Close()
	// 读取私钥
	fileInfo, _ := private.Stat()
	buf := make([]byte,fileInfo.Size())
	_, _ = private.Read(buf)
	// pem解码
	block, _ := pem.Decode(buf)
	// 转换成私钥结构体
	privateKey,_ := x509.ParsePKCS1PrivateKey(block.Bytes)
	// 解密密文
	plainText,err2 := rsa.DecryptPKCS1v15(nil,privateKey,src)
	return string(plainText),err2
}