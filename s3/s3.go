package s3

import (
	
	"github.com/goamz/goamz/aws"
	"github.com/goamz/goamz/s3"
	"log"
)

var (
	AccessKeyId, SecretAccessKey, Bucket, Region string
	Expire                                       int
)

//us-east-1, us-west-1, us-west-2, eu-west-1, eu-central-1, ap-southeast-1, ap-southeast-2, ap-northeast-1, sa-east-1, cn-north-1

var client *s3.S3

func auth() {
	auth := aws.Auth{AccessKey: AccessKeyId, SecretKey: SecretAccessKey}
	if Region == "google" {
		client = s3.New(auth, aws.Region{Name: "gs", S3Endpoint: "http://storage.googleapis.com"})
	} else {

		client = s3.New(auth, aws.Regions[Region])

	}

}

func Exists(object string) bool {
	auth()
	b := client.Bucket(Bucket)
	exists, _ := b.Exists(object)

	return exists
}

func Del(object string) {
	auth()
	b := client.Bucket(Bucket)
	b.Del(object)
}

func Get(object string) ([]byte, error) {

	auth()

	b := client.Bucket(Bucket)
	data, err := b.Get(object)

	if err != nil {
		log.Println("failed to get object :" + object + ": from bucket " + Bucket)
		log.Println(err)
		return make([]byte, 0), err
	} else {
		return data, nil
	}
}

func Put(data []byte, path string, ctype string) error {
	auth()
	b := client.Bucket(Bucket)

	// err := b.Put(path, data, helper.GetImageType(ctype), s3.PublicRead)

	param := map[string][]string{
		"Content-Type": {GetImageType(ctype)},
	}
	if Region != "google" {
		param["x-amz-storage-class"] = []string{"REDUCED_REDUNDANCY"}
	}

	err := b.PutHeader(path, data, param, s3.PublicRead)

	if err != nil {
		log.Println("failed to store object :" + path + ":" + " on :" + Bucket)
		log.Println(err)
		return err
	} else {
		log.Println("succesfully stored :" + path + ":" + " on :" + Bucket)
		return nil
	}
}

func GetImageType(ext string) string {
	switch ext {
	case "jpeg":
		return "image/jpeg"
	case "png":
		return "image/png"
	case "webp":
		return "image/webp"
	}
	return ""
}